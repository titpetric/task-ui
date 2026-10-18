package workflow_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite" // the driver storage opens

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
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

	run, err := f.runner.Start(context.Background(), wf, ops)
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

	run, err := f.runner.Start(context.Background(), wf, ops)
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

	run, err := f.runner.Start(context.Background(), wf, ops)
	assert.NoError(t, err)
	_, err = f.runner.Start(context.Background(), wf, ops)
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

	_, err := f.runner.Start(context.Background(), wf, ops)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "infra / gone")
	runs, _ := f.wfRuns.ListByWorkflow(context.Background(), wf.ID, 10)
	assert.Len(t, runs, 0)
}
