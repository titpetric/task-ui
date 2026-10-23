//go:build integration

package workflow_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// prompter is a Prompter that records each prompt, and the step states
// before it at the time, and returns err, or waits for a stop with block.
type prompter struct {
	mu     chan struct{}
	asked  []string
	states []string
	err    error
	block  bool
	wfRuns *storage.WorkflowRunStorage
}

// AskStep records a prompt or summary step and the steps before it, and
// answers with the prompt, or "summary" for a summary.
func (p *prompter) AskStep(ctx context.Context, user *model.User, run *model.WorkflowRun, step *model.WorkflowRunStep, place, input string) (string, string, error) {
	steps, _ := p.wfRuns.Steps(context.Background(), run.ID)
	what := step.Prompt
	if step.IsSummary() {
		what = "summary"
	}
	p.mu <- struct{}{}
	p.asked = append(p.asked, user.Name+": "+what+" in "+place+" / "+input)
	for _, prior := range steps {
		if prior.Position >= step.Position {
			break
		}
		p.states = append(p.states, prior.State)
		if prior.Answer != "" {
			p.states = append(p.states, prior.Answer)
		}
	}
	<-p.mu
	if p.block {
		<-ctx.Done()
		return "", "", ctx.Err()
	}
	return "r1", "the answer to " + what, p.err
}

func newPrompter(f *fixture) *prompter {
	p := &prompter{mu: make(chan struct{}, 1), wfRuns: f.wfRuns}
	f.runner.Prompter = p
	return p
}

// asks is a prompt step in the root, summary a summary step there.
func asks(prompt string) model.WorkflowStep {
	return model.WorkflowStep{Kind: model.StepPrompt, Folder: "_", Job: model.PromptJob, Prompt: prompt}
}

var summary = model.WorkflowStep{Kind: model.StepSummary, Folder: "_", Job: model.PromptJob}

// A last prompt step runs once the steps before it ended, the run still
// running, with the run's input, and its answer is the run's. A prompt
// alone is the whole workflow; a failing prompt fails the run.
func TestRunner_Prompt(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	p := newPrompter(f)
	wf := &model.Workflow{Name: "summed"}
	assert.NoError(t, f.workflows.Create(ctx, wf, []model.WorkflowStep{{Folder: "infra", Job: "build"}, {Folder: "infra", Job: "test", Disabled: true}, asks("summarize")}))
	run, err := f.runner.Start(ctx, wf, ops, "release 1.2")
	assert.NoError(t, err)
	ended := f.wait(t, run.ID)
	assert.Equal(t, model.StateOK, ended.State)
	assert.Equal(t, "the answer to summarize", ended.Answer)
	assert.Equal(t, []string{"ops: summarize in _ / release 1.2"}, p.asked)
	assert.Equal(t, []string{"ok", "skipped"}, p.states)

	alone := &model.Workflow{Name: "alone"}
	assert.NoError(t, f.workflows.Create(ctx, alone, []model.WorkflowStep{asks("say ok")}))
	run, err = f.runner.Start(ctx, alone, ops, "")
	assert.NoError(t, err)
	assert.Equal(t, model.StateOK, f.wait(t, run.ID).State)
	assert.Equal(t, "ops: say ok in _ / ", p.asked[1])

	p.err = errors.New("no answer")
	run, err = f.runner.Start(ctx, alone, ops, "")
	assert.NoError(t, err)
	failed := f.wait(t, run.ID)
	assert.Equal(t, model.StateFailed, failed.State)
	assert.Equal(t, "the prompt in _ no answer", failed.Message)

	// Without a Prompter the prompt fails, and the run.
	f.runner.Prompter = nil
	run, err = f.runner.Start(ctx, alone, ops, "")
	assert.NoError(t, err)
	assert.Equal(t, "the prompt in _ prompts aren't available on this instance", f.wait(t, run.ID).Message)
}

// After a failed step the run skips a prompt step, but a summary step runs
// and asks about the failure; the run keeps the step's failure and the
// summary's answer.
func TestRunner_SummaryAfterFailure(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	p := newPrompter(f)
	broken := &model.Workflow{Name: "broken"}
	assert.NoError(t, f.workflows.Create(ctx, broken, []model.WorkflowStep{{Folder: "infra", Job: "broken"}, asks("why"), summary}))
	run, err := f.runner.Start(ctx, broken, ops, "")
	assert.NoError(t, err)
	ended := f.wait(t, run.ID)
	assert.Equal(t, model.StateFailed, ended.State)
	assert.Contains(t, ended.Message, "infra / broken")
	assert.Equal(t, "the answer to summary", ended.Answer)
	assert.Equal(t, []string{"ops: summary in _ / "}, p.asked)
	assert.Equal(t, []string{"failed", "skipped"}, p.states)
	steps, _ := f.wfRuns.Steps(ctx, run.ID)
	assert.Equal(t, model.StepSkipped, steps[1].State)
	assert.Equal(t, model.StateOK, steps[2].State)
	assert.Equal(t, model.StepSummary, steps[2].Kind)
	assert.Equal(t, "the answer to summary", steps[2].Answer)
}

// A prompt step asks in its place with the run's input and ends ok with
// Claude's answer, which the steps after it read; a failing prompt step
// fails the run and skips the rest, a summary after it aside.
func TestRunner_PromptStep(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	p := newPrompter(f)
	ask := asks("fetch the weather")
	wf := &model.Workflow{Name: "weather"}
	assert.NoError(t, f.workflows.Create(ctx, wf, []model.WorkflowStep{ask, {Folder: "infra", Job: "build"}, asks("water today?")}))
	run, err := f.runner.Start(ctx, wf, ops, "porec")
	assert.NoError(t, err)
	assert.Equal(t, model.StateOK, f.wait(t, run.ID).State)
	assert.Equal(t, []string{"ops: fetch the weather in _ / porec", "ops: water today? in _ / porec"}, p.asked)
	assert.Equal(t, []string{"ok", "the answer to fetch the weather", "ok"}, p.states)
	steps, _ := f.wfRuns.Steps(ctx, run.ID)
	assert.Equal(t, "r1", steps[0].RunID)
	assert.Equal(t, "fetch the weather", steps[0].Prompt)
	assert.Equal(t, model.StepPrompt, steps[0].Kind)
	ended, _ := f.wfRuns.Get(ctx, run.ID)
	assert.Equal(t, "porec", ended.Input)
	assert.Equal(t, "the answer to water today?", ended.Answer, "the last answer is the run's")

	p.err = errors.New("no answer")
	run, err = f.runner.Start(ctx, wf, ops, "")
	assert.NoError(t, err)
	failed := f.wait(t, run.ID)
	assert.Equal(t, model.StateFailed, failed.State)
	assert.Equal(t, "the prompt in _ no answer", failed.Message)
	assert.Equal(t, "", failed.Answer)
	steps, _ = f.wfRuns.Steps(ctx, run.ID)
	assert.Equal(t, model.StateFailed, steps[0].State)
	assert.Equal(t, "", steps[0].Answer)
	assert.Equal(t, model.StepSkipped, steps[1].State)
	assert.Equal(t, model.StepSkipped, steps[2].State)
}

// A stop while a prompt runs stops it, and the run; the run skips a
// summary after it.
func TestRunner_StopPrompt(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	p := newPrompter(f)
	p.block = true
	wf := &model.Workflow{Name: "waits"}
	assert.NoError(t, f.workflows.Create(ctx, wf, []model.WorkflowStep{asks("say ok"), summary}))
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
	assert.Equal(t, "stopped at the prompt in _", ended.Message)
	steps, _ := f.wfRuns.Steps(ctx, run.ID)
	assert.Equal(t, model.StepSkipped, steps[1].State)
}
