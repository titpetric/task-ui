package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func steps(pairs ...string) []model.WorkflowStep {
	out := []model.WorkflowStep{}
	for i := 0; i+1 < len(pairs); i += 2 {
		out = append(out, model.WorkflowStep{Folder: pairs[i], Job: pairs[i+1]})
	}
	return out
}

func TestWorkflowStorage_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	workflows := storage.NewWorkflowStorage(newDB(t))

	wf := &model.Workflow{Name: "deploy", UserName: "ops"}
	assert.NoError(t, workflows.Create(ctx, wf, steps("dockerfiles/task", "build", "dockerfiles/task", "push", "deploy/terraform", "apply")))
	assert.True(t, model.ValidID(wf.ID))

	got, err := workflows.Steps(ctx, wf.ID)
	assert.NoError(t, err)
	assert.Len(t, got, 3)
	assert.Equal(t, "build", got[0].Job)
	assert.Equal(t, "apply", got[2].Job)
	assert.Equal(t, int64(2), got[2].Position)

	assert.Error(t, workflows.Create(ctx, &model.Workflow{Name: "empty"}, nil), "no steps")
	assert.Error(t, workflows.Create(ctx, &model.Workflow{}, steps("a", "b")), "no name")
	assert.Error(t, workflows.Create(ctx, &model.Workflow{Name: "half"}, steps("a", "")), "a step without a job")
}

func TestWorkflowStorage_GetListUpdateDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	workflows := storage.NewWorkflowStorage(newDB(t))

	b := &model.Workflow{Name: "b release"}
	a := &model.Workflow{Name: "a build"}
	assert.NoError(t, workflows.Create(ctx, b, steps("x", "one")))
	assert.NoError(t, workflows.Create(ctx, a, steps("x", "two")))

	list, err := workflows.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "a build", list[0].Name)

	got, err := workflows.Get(ctx, b.ID)
	assert.NoError(t, err)
	assert.Equal(t, "b release", got.Name)
	_, err = workflows.Get(ctx, "missing")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// An update replaces the name and the whole chain.
	got.Name = "b release v2"
	assert.NoError(t, workflows.Update(ctx, got, steps("y", "test", "y", "push")))
	chain, err := workflows.Steps(ctx, b.ID)
	assert.NoError(t, err)
	assert.Len(t, chain, 2)
	assert.Equal(t, "test", chain[0].Job)
	assert.Error(t, workflows.Update(ctx, got, nil), "a workflow keeps at least one step")

	assert.NoError(t, workflows.Delete(ctx, b.ID))
	_, err = workflows.Get(ctx, b.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	chain, err = workflows.Steps(ctx, b.ID)
	assert.NoError(t, err)
	assert.Len(t, chain, 0)
}

func TestWorkflowRunStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	runs := storage.NewWorkflowRunStorage(db)

	run := &model.WorkflowRun{WorkflowID: "wf", Name: "deploy", UserName: "ops"}
	assert.NoError(t, runs.Create(ctx, run, steps("a", "build", "a", "push")))
	assert.Equal(t, model.StateRunning, run.State)

	chain, err := runs.Steps(ctx, run.ID)
	assert.NoError(t, err)
	assert.Len(t, chain, 2)
	assert.Equal(t, model.StepPending, chain[0].State)

	chain[0].RunID, chain[0].State = "run1", model.StateOK
	assert.NoError(t, runs.UpdateStep(ctx, &chain[0]))
	chain, _ = runs.Steps(ctx, run.ID)
	assert.Equal(t, "run1", chain[0].RunID)

	run.State, run.Message = model.StateFailed, "push failed"
	assert.NoError(t, runs.Update(ctx, run))
	got, err := runs.Get(ctx, run.ID)
	assert.NoError(t, err)
	assert.Equal(t, "push failed", got.Message)

	list, err := runs.ListByWorkflow(ctx, "wf", 10)
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	// A restart ends what was running: the run fails, a running step with
	// it, and steps that never started are skipped.
	open := &model.WorkflowRun{WorkflowID: "wf", Name: "deploy"}
	assert.NoError(t, runs.Create(ctx, open, steps("a", "build", "a", "push")))
	openSteps, _ := runs.Steps(ctx, open.ID)
	openSteps[0].State = model.StateRunning
	assert.NoError(t, runs.UpdateStep(ctx, &openSteps[0]))
	assert.NoError(t, runs.Interrupt(ctx, "task-ui restarted"))
	ended, _ := runs.Get(ctx, open.ID)
	assert.Equal(t, model.StateFailed, ended.State)
	assert.Equal(t, "task-ui restarted", ended.Message)
	openSteps, _ = runs.Steps(ctx, open.ID)
	assert.Equal(t, model.StateFailed, openSteps[0].State)
	assert.Equal(t, model.StepSkipped, openSteps[1].State)
}
