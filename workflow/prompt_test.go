package workflow_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
)

// prompter is a Prompter that records each prompt and the step states
// at the time, and returns err, or waits for a stop with block.
type prompter struct {
	mu     chan struct{}
	asked  []string
	states []string
	err    error
	block  bool
	wfRuns *storage.WorkflowRunStorage
}

func (p *prompter) AskWorkflow(ctx context.Context, user *model.User, run *model.WorkflowRun, wf *model.Workflow, input string) (string, error) {
	steps, _ := p.wfRuns.Steps(context.Background(), run.ID)
	p.mu <- struct{}{}
	p.asked = append(p.asked, user.Name+": "+wf.Prompt+" / "+input+" / "+run.State)
	for _, step := range steps {
		p.states = append(p.states, step.State)
		if step.Answer != "" {
			p.states = append(p.states, step.Answer)
		}
	}
	<-p.mu
	if p.block {
		<-ctx.Done()
		return "", ctx.Err()
	}
	return "the answer to " + wf.Prompt, p.err
}

// AskStep records a prompt step as AskWorkflow does the workflow's, and
// answers with the prompt and its place.
func (p *prompter) AskStep(ctx context.Context, user *model.User, run *model.WorkflowRun, step *model.WorkflowRunStep, place, input string) (string, string, error) {
	p.mu <- struct{}{}
	p.asked = append(p.asked, user.Name+": step "+step.Prompt+" in "+place+" / "+input)
	<-p.mu
	if p.block {
		<-ctx.Done()
		return "", "", ctx.Err()
	}
	return "r1", "the answer to " + step.Prompt, p.err
}

func newPrompter(f *fixture) *prompter {
	p := &prompter{mu: make(chan struct{}, 1), wfRuns: f.wfRuns}
	f.runner.Prompter = p
	return p
}

// A workflow with a prompt runs it once the steps ended, the run still
// running, with the run's input; the run ends with the steps' state. A
// prompt alone is the whole workflow; a failing prompt fails a run
// whose steps ended ok.
func TestRunner_Prompt(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	p := newPrompter(f)
	wf := &model.Workflow{Name: "summed", Prompt: "summarize"}
	assert.NoError(t, f.workflows.Create(ctx, wf, []model.WorkflowStep{{Folder: "infra", Job: "build"}, {Folder: "infra", Job: "test", Disabled: true}}))
	run, err := f.runner.Start(ctx, wf, ops, "release 1.2")
	assert.NoError(t, err)
	assert.Equal(t, model.StateOK, f.wait(t, run.ID).State)
	assert.Equal(t, []string{"ops: summarize / release 1.2 / ok"}, p.asked)
	assert.Equal(t, []string{"ok", "skipped"}, p.states)

	alone := &model.Workflow{Name: "alone", Prompt: "say ok"}
	assert.NoError(t, f.workflows.Create(ctx, alone, nil))
	run, err = f.runner.Start(ctx, alone, ops, "")
	assert.NoError(t, err)
	assert.Equal(t, model.StateOK, f.wait(t, run.ID).State)
	assert.Equal(t, "ops: say ok /  / ok", p.asked[1])

	p.err = errors.New("no answer")
	run, err = f.runner.Start(ctx, alone, ops, "")
	assert.NoError(t, err)
	failed := f.wait(t, run.ID)
	assert.Equal(t, model.StateFailed, failed.State)
	assert.Equal(t, "the prompt: no answer", failed.Message)

	// Failed steps keep their state and message, and the prompt still
	// asks about them.
	broken := &model.Workflow{Name: "broken", Prompt: "why"}
	assert.NoError(t, f.workflows.Create(ctx, broken, []model.WorkflowStep{{Folder: "infra", Job: "broken"}}))
	run, err = f.runner.Start(ctx, broken, ops, "")
	assert.NoError(t, err)
	ended := f.wait(t, run.ID)
	assert.Equal(t, model.StateFailed, ended.State)
	assert.Contains(t, ended.Message, "infra / broken")
	assert.Len(t, p.asked, 4)

	// Without a Prompter the prompt does not run, and the run fails.
	f.runner.Prompter = nil
	run, err = f.runner.Start(ctx, alone, ops, "")
	assert.NoError(t, err)
	assert.Equal(t, "the prompt: prompts aren't available on this instance", f.wait(t, run.ID).Message)
}

// A prompt step asks in its place with the run's input and ends ok with
// Claude's answer, which the steps after it read; a failing prompt step
// fails the run and skips the rest.
func TestRunner_PromptStep(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	p := newPrompter(f)
	ask := model.WorkflowStep{Folder: "_", Job: model.PromptJob, Prompt: "fetch the weather"}
	wf := &model.Workflow{Name: "weather", Prompt: "water today?"}
	assert.NoError(t, f.workflows.Create(ctx, wf, []model.WorkflowStep{ask, {Folder: "infra", Job: "build"}}))
	run, err := f.runner.Start(ctx, wf, ops, "porec")
	assert.NoError(t, err)
	assert.Equal(t, model.StateOK, f.wait(t, run.ID).State)
	assert.Equal(t, []string{"ops: step fetch the weather in _ / porec", "ops: water today? / porec / ok"}, p.asked)
	assert.Equal(t, []string{"ok", "the answer to fetch the weather", "ok"}, p.states)
	steps, _ := f.wfRuns.Steps(ctx, run.ID)
	assert.Equal(t, "r1", steps[0].RunID)
	assert.Equal(t, "fetch the weather", steps[0].Prompt)

	p.err = errors.New("no answer")
	run, err = f.runner.Start(ctx, wf, ops, "")
	assert.NoError(t, err)
	failed := f.wait(t, run.ID)
	assert.Equal(t, model.StateFailed, failed.State)
	assert.Equal(t, "the prompt in _ no answer", failed.Message)
	steps, _ = f.wfRuns.Steps(ctx, run.ID)
	assert.Equal(t, model.StateFailed, steps[0].State)
	assert.Equal(t, "", steps[0].Answer)
	assert.Equal(t, model.StepSkipped, steps[1].State)
}

// A stop while the prompt runs stops it, and the run.
func TestRunner_StopPrompt(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	p := newPrompter(f)
	p.block = true
	wf := &model.Workflow{Name: "waits", Prompt: "say ok"}
	assert.NoError(t, f.workflows.Create(ctx, wf, nil))
	run, err := f.runner.Start(ctx, wf, ops, "")
	assert.NoError(t, err)
	assert.Eventually(t, func() bool {
		p.mu <- struct{}{}
		defer func() { <-p.mu }()
		return len(p.asked) == 1
	}, 5*time.Second, 10*time.Millisecond)
	assert.NoError(t, f.runner.Stop(run.ID))
	ended := f.wait(t, run.ID)
	assert.Equal(t, model.StateStopped, ended.State)
	assert.Equal(t, "stopped at the prompt", ended.Message)
}

// The notes board runs a workflow's prompt.
func TestPrompter(t *testing.T) {
	t.Parallel()
	var p workflow.Prompter = &note.Board{}
	assert.True(t, p != nil)
}
