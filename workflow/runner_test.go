package workflow_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite" // the driver storage opens

	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
)

// catalogue is a fixed set of folders.
type catalogue map[string]*folder.Folder

func (c catalogue) Find(id string) *folder.Folder { return c[id] }

type fixture struct {
	runner    *workflow.Runner
	workflows *storage.WorkflowStorage
	wfRuns    *storage.WorkflowRunStorage
	runs      *storage.RunStorage
}

func job(id, script string) *folder.Job {
	j := &folder.Job{Folder: "infra", ID: id, Command: []string{"sh", "-c", script}}
	j.Name = id
	return j
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	dir := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(dir, "infra"), 0o755))
	infra := &folder.Folder{ID: "infra", Dir: filepath.Join(dir, "infra")}
	infra.Jobs = []*folder.Job{
		job("build", "echo built"),
		job("test", "echo tested"),
		job("broken", "echo broken; exit 3"),
		job("slow", "sleep 30"),
	}

	runs := storage.NewRunStorage(db)
	manager := runner.New(runner.Options{Runs: runs, Recordings: &recording.Store{Root: t.TempDir()}, Grace: time.Second})
	t.Cleanup(func() { manager.Shutdown(context.Background()) })

	f := &fixture{workflows: storage.NewWorkflowStorage(db), wfRuns: storage.NewWorkflowRunStorage(db), runs: runs}
	f.runner = workflow.New(f.workflows, f.wfRuns, catalogue{"infra": infra}, manager)
	return f
}

func (f *fixture) create(t *testing.T, jobs ...string) *model.Workflow {
	t.Helper()
	steps := []model.WorkflowStep{}
	for _, j := range jobs {
		steps = append(steps, model.WorkflowStep{Folder: "infra", Job: j})
	}
	wf := &model.Workflow{Name: "chain"}
	assert.NoError(t, f.workflows.Create(context.Background(), wf, steps))
	return wf
}

func (f *fixture) wait(t *testing.T, id string) *model.WorkflowRun {
	t.Helper()
	select {
	case <-f.runner.Done(id):
	case <-time.After(20 * time.Second):
		t.Fatal("the workflow run did not finish")
	}
	run, err := f.wfRuns.Get(context.Background(), id)
	assert.NoError(t, err)
	return run
}

var ops = &model.User{ID: model.NewID(), Name: "ops"}

func TestRunner_StepsInOrder(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	wf := f.create(t, "build", "test")

	run, err := f.runner.Start(context.Background(), wf, ops, "")
	assert.NoError(t, err)
	assert.Equal(t, model.StateRunning, run.State)
	assert.Equal(t, "chain", run.Name)
	assert.Equal(t, "ops", run.UserName)

	ended := f.wait(t, run.ID)
	assert.Equal(t, model.StateOK, ended.State)
	steps, err := f.wfRuns.Steps(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.Len(t, steps, 2)
	for _, step := range steps {
		assert.Equal(t, model.StateOK, step.State)
		assert.True(t, step.RunID != "")
	}
	first, _ := f.runs.Get(context.Background(), steps[0].RunID)
	second, _ := f.runs.Get(context.Background(), steps[1].RunID)
	assert.Equal(t, "build", first.Job)
	assert.Equal(t, "ops", second.UserName)
	assert.True(t, !second.StartedAt.Before(*first.FinishedAt), "the second step starts after the first ended")
}

func TestRunner_StopsAtAFailure(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	wf := f.create(t, "build", "broken", "test")

	run, err := f.runner.Start(context.Background(), wf, ops, "")
	assert.NoError(t, err)
	ended := f.wait(t, run.ID)
	assert.Equal(t, model.StateFailed, ended.State)
	assert.Contains(t, ended.Message, "infra / broken")

	steps, _ := f.wfRuns.Steps(context.Background(), run.ID)
	assert.Equal(t, model.StateOK, steps[0].State)
	assert.Equal(t, model.StateFailed, steps[1].State)
	assert.Equal(t, model.StepSkipped, steps[2].State)
	assert.Equal(t, "", steps[2].RunID)
}

func TestRunner_OneRunAtATime(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	wf := f.create(t, "slow", "build")

	run, err := f.runner.Start(context.Background(), wf, ops, "")
	assert.NoError(t, err)
	_, err = f.runner.Start(context.Background(), wf, ops, "")
	assert.ErrorIs(t, err, workflow.ErrBusy)
	active, ok := f.runner.Active(wf.ID)
	assert.True(t, ok)
	assert.Equal(t, run.ID, active)

	assert.NoError(t, f.runner.Stop(run.ID))
	ended := f.wait(t, run.ID)
	steps, _ := f.wfRuns.Steps(context.Background(), run.ID)
	assert.Equal(t, model.StateStopped, ended.State)
	assert.Equal(t, model.StateStopped, steps[0].State)
	assert.Equal(t, model.StepSkipped, steps[1].State)
	assert.ErrorIs(t, f.runner.Stop(run.ID), workflow.ErrNotRunning)
}

func TestRunner_MissingJob(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	wf := f.create(t, "build", "gone")

	_, err := f.runner.Start(context.Background(), wf, ops, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "infra / gone")
	runs, _ := f.wfRuns.ListByWorkflow(context.Background(), wf.ID, storage.WorkflowRunFilter{Limit: 10})
	assert.Len(t, runs, 0)
}

// A step that names a workflow runs it, as a run of its own, and waits
// for it.
func TestRunner_NestedWorkflow(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	inner := f.create(t, "build")
	outer := &model.Workflow{Name: "outer"}
	assert.NoError(t, f.workflows.Create(ctx, outer, []model.WorkflowStep{{Folder: "infra", Job: "test"}, {WorkflowRef: inner.ID}}))

	run, err := f.runner.Start(ctx, outer, ops, "")
	assert.NoError(t, err)
	ended := f.wait(t, run.ID)
	assert.Equal(t, model.StateOK, ended.State)

	steps, err := f.wfRuns.Steps(ctx, run.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.StateOK, steps[1].State)
	assert.Equal(t, "", steps[1].RunID)
	assert.True(t, steps[1].WorkflowRunRef != "", "the run of inner")
	innerRun, err := f.wfRuns.Get(ctx, steps[1].WorkflowRunRef)
	assert.NoError(t, err)
	assert.Equal(t, inner.ID, innerRun.WorkflowID)
	assert.Equal(t, model.StateOK, innerRun.State)
	assert.Equal(t, "ops", innerRun.UserName)
	innerSteps, _ := f.wfRuns.Steps(ctx, innerRun.ID)
	assert.True(t, innerSteps[0].RunID != "")
	first, _ := f.runs.Get(ctx, steps[0].RunID)
	second, _ := f.runs.Get(ctx, innerSteps[0].RunID)
	assert.True(t, !second.StartedAt.Before(*first.FinishedAt), "inner starts after the step before it ended")
}

// A failing nested workflow fails the step that ran it, and the runner
// skips the rest.
func TestRunner_NestedFailure(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	inner := f.create(t, "broken")
	outer := &model.Workflow{Name: "outer"}
	assert.NoError(t, f.workflows.Create(ctx, outer, []model.WorkflowStep{{WorkflowRef: inner.ID}, {Folder: "infra", Job: "test"}}))

	run, err := f.runner.Start(ctx, outer, ops, "")
	assert.NoError(t, err)
	ended := f.wait(t, run.ID)
	assert.Equal(t, model.StateFailed, ended.State)
	assert.Contains(t, ended.Message, "chain")
	steps, _ := f.wfRuns.Steps(ctx, run.ID)
	assert.Equal(t, model.StateFailed, steps[0].State)
	assert.Equal(t, model.StepSkipped, steps[1].State)
}

// Stopping a workflow stops the workflow its step is running.
func TestRunner_StopNested(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	inner := f.create(t, "slow")
	outer := &model.Workflow{Name: "outer"}
	assert.NoError(t, f.workflows.Create(ctx, outer, []model.WorkflowStep{{WorkflowRef: inner.ID}}))

	run, err := f.runner.Start(ctx, outer, ops, "")
	assert.NoError(t, err)
	deadline := time.Now().Add(10 * time.Second)
	innerRun, started := "", false
	for !started && time.Now().Before(deadline) {
		innerRun, started = f.runner.Active(inner.ID)
		time.Sleep(20 * time.Millisecond)
	}
	assert.True(t, started, "inner started")
	assert.NoError(t, f.runner.Stop(run.ID))
	ended := f.wait(t, run.ID)
	assert.Equal(t, model.StateStopped, ended.State)
	innerEnded := f.wait(t, innerRun)
	assert.Equal(t, model.StateStopped, innerEnded.State)
}

// A run reports on the workflow's topic: the run, each step as it
// starts and ends, and the end.
func TestRunner_Events(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.runner.Events = event.NewBus()
	wf := f.create(t, "build", "test")
	ch, cancel := f.runner.Events.Subscribe("workflow:" + wf.ID)
	defer cancel()

	run, err := f.runner.Start(context.Background(), wf, ops, "")
	assert.NoError(t, err)
	f.wait(t, run.ID)

	names, states := []string{}, []string{}
	for len(names) == 0 || names[len(names)-1] != "done" {
		select {
		case e := <-ch:
			names = append(names, e.Name)
			if step, ok := e.Data.(workflow.StepEvent); ok {
				states = append(states, step.State)
				assert.Equal(t, run.ID, step.RunID)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("events so far: %v", names)
		}
	}
	assert.Equal(t, []string{"run", "step", "step", "step", "step", "done"}, names)
	assert.Equal(t, []string{"running", "ok", "running", "ok"}, states)
}

// Resolve finds every step's folder and job, or its workflow, and says
// which step is missing.
func TestRunner_Resolve(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	inner := f.create(t, "build")

	var steps []workflow.Step
	steps, err := f.runner.Resolve(ctx, []model.WorkflowStep{{Folder: "infra", Job: "test"}, {WorkflowRef: inner.ID}})
	assert.NoError(t, err)
	assert.Len(t, steps, 2)
	assert.Equal(t, "infra", steps[0].Folder.ID)
	assert.Equal(t, "test", steps[0].Job.ID)
	assert.True(t, steps[0].Workflow == nil)
	assert.Equal(t, inner.ID, steps[1].Workflow.ID)
	assert.True(t, steps[1].Job == nil)

	_, err = f.runner.Resolve(ctx, []model.WorkflowStep{{Folder: "infra", Job: "gone"}})
	assert.ErrorContains(t, err, "no job infra / gone")
	_, err = f.runner.Resolve(ctx, []model.WorkflowStep{{Folder: "nowhere", Job: "build"}})
	assert.ErrorContains(t, err, "no job nowhere / build")
	_, err = f.runner.Resolve(ctx, []model.WorkflowStep{{WorkflowRef: model.NewID()}})
	assert.ErrorContains(t, err, "no workflow")
	_, err = f.runner.Resolve(ctx, []model.WorkflowStep{{Folder: "infra", Job: folder.ShellID}})
	assert.ErrorIs(t, err, folder.ErrShellUnattended)
}

// A step is a job of a folder, or a workflow in its place.
func TestStep(t *testing.T) {
	t.Parallel()
	j := job("build", "echo built")
	s := workflow.Step{Folder: &folder.Folder{ID: "infra"}, Job: j}
	assert.Equal(t, "build", s.Job.ID)
	assert.True(t, workflow.Step{Workflow: &model.Workflow{ID: "w1"}}.Job == nil)
}

// Anything that finds a folder by ID is a catalogue: the discovery, or
// a fixed set of folders.
func TestCatalogue(t *testing.T) {
	t.Parallel()
	var c workflow.Catalogue = catalogue{"infra": {ID: "infra"}}
	assert.Equal(t, "infra", c.Find("infra").ID)
	assert.True(t, c.Find("gone") == nil)
	var _ workflow.Catalogue = &folder.Discovery{}
}

// The run and done events carry the run as the browser reads it.
func TestRunEvent(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(workflow.RunEvent{RunID: "wr1", WorkflowID: "w1", Name: "release", State: "ok", UserName: "ops"})
	assert.NoError(t, err)
	assert.Equal(t, `{"run_id":"wr1","workflow_id":"w1","name":"release","state":"ok","user_name":"ops"}`, string(data))
}

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

func (p *prompter) AskWorkflow(ctx context.Context, user *model.User, run *model.WorkflowRun, wf *model.Workflow, input string) error {
	steps, _ := p.wfRuns.Steps(context.Background(), run.ID)
	p.mu <- struct{}{}
	p.asked = append(p.asked, user.Name+": "+wf.Prompt+" / "+input+" / "+run.State)
	for _, step := range steps {
		p.states = append(p.states, step.State)
	}
	<-p.mu
	if p.block {
		<-ctx.Done()
		return ctx.Err()
	}
	return p.err
}

func newPrompter(f *fixture) *prompter {
	p := &prompter{mu: make(chan struct{}, 1), wfRuns: f.wfRuns}
	f.runner.Prompter = p
	return p
}

// The runner skips a disabled step and runs the steps after it; a
// workflow with every step disabled and no prompt does not start.
func TestRunner_DisabledStep(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	wf := &model.Workflow{Name: "toggled"}
	// A disabled step does not stop the run.
	assert.NoError(t, f.workflows.Create(ctx, wf, []model.WorkflowStep{
		{Folder: "infra", Job: "build"}, {Folder: "infra", Job: "broken", Disabled: true}, {Folder: "infra", Job: "test"},
	}))
	run, err := f.runner.Start(ctx, wf, ops, "")
	assert.NoError(t, err)
	ended := f.wait(t, run.ID)
	assert.Equal(t, model.StateOK, ended.State)
	steps, _ := f.wfRuns.Steps(ctx, run.ID)
	assert.Equal(t, model.StateOK, steps[0].State)
	assert.Equal(t, model.StepSkipped, steps[1].State)
	assert.Equal(t, "", steps[1].RunID)
	assert.Equal(t, model.StateOK, steps[2].State)

	// Storage keeps such a workflow out; the runner refuses one too.
	off := &model.Workflow{Name: "off", Prompt: "say ok"}
	assert.NoError(t, f.workflows.Create(ctx, off, []model.WorkflowStep{{Folder: "infra", Job: "build", Disabled: true}}))
	off.Prompt = ""
	_, err = f.runner.Start(ctx, off, ops, "")
	assert.ErrorIs(t, err, workflow.ErrNothing)
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
