package storage_test

import (
	"context"
	"testing"
	"time"

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
	off := []model.WorkflowStep{{Folder: "a", Job: "b", Disabled: true}}
	assert.Error(t, workflows.Create(ctx, &model.Workflow{Name: "off"}, off), "every step disabled, no prompt")

	// A prompt runs a workflow with every step disabled, or with none,
	// and the prompt, its place and the step toggles round trip.
	asks := &model.Workflow{Name: "asks", Prompt: "summarize, change nothing", PromptFolder: "malta/dockerfiles"}
	assert.NoError(t, workflows.Create(ctx, asks, append(off, model.WorkflowStep{Folder: "a", Job: "c"})))
	read, err := workflows.Get(ctx, asks.ID)
	assert.NoError(t, err)
	assert.Equal(t, "summarize, change nothing", read.Prompt)
	assert.Equal(t, "malta/dockerfiles", read.PromptFolder)
	toggled, err := workflows.Steps(ctx, asks.ID)
	assert.NoError(t, err)
	assert.True(t, toggled[0].Disabled)
	assert.False(t, toggled[1].Disabled)
	alone := &model.Workflow{Name: "alone", Prompt: "say ok"}
	assert.NoError(t, workflows.Create(ctx, alone, nil))
	none, err := workflows.Steps(ctx, alone.ID)
	assert.NoError(t, err)
	assert.Len(t, none, 0)
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

	list, err := runs.ListByWorkflow(ctx, "wf", storage.WorkflowRunFilter{Limit: 10})
	assert.NoError(t, err)
	assert.Len(t, list, 1)

	// The runs of every workflow, newest first, for the sidebar. Its ID
	// sorts later, for a start in the same millisecond.
	other := &model.WorkflowRun{ID: model.NewIDAt(time.Now().Add(time.Second)), WorkflowID: "wf2", Name: "release", UserName: "ops"}
	assert.NoError(t, runs.Create(ctx, other, steps("a", "build")))
	all, err := runs.List(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, all, 2)
	assert.Equal(t, "release", all[0].Name)
	one, _ := runs.List(ctx, 1)
	assert.Len(t, one, 1)

	// A restart ends the open runs: the run fails, a running step with
	// it, and Interrupt skips steps that never started.
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

// A step runs another workflow in place of a job. A workflow may not
// reach itself through the workflows it runs, and one in use stays.
func TestWorkflowStorage_WorkflowSteps(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	workflows := storage.NewWorkflowStorage(newDB(t))

	inner := &model.Workflow{Name: "inner"}
	assert.NoError(t, workflows.Create(ctx, inner, steps("infra", "build")))
	outer := &model.Workflow{Name: "outer"}
	assert.NoError(t, workflows.Create(ctx, outer, []model.WorkflowStep{{Folder: "infra", Job: "test"}, {WorkflowRef: inner.ID}}))
	got, err := workflows.Steps(ctx, outer.ID)
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, inner.ID, got[1].WorkflowRef)
	assert.Equal(t, "", got[1].Folder)

	assert.Error(t, workflows.Create(ctx, &model.Workflow{Name: "both"}, []model.WorkflowStep{{Folder: "infra", Job: "build", WorkflowRef: inner.ID}}), "a job or a workflow, not both")
	assert.Error(t, workflows.Create(ctx, &model.Workflow{Name: "gone"}, []model.WorkflowStep{{WorkflowRef: model.NewID()}}), "a workflow that is not there")

	// inner may not run outer, which runs inner, nor itself.
	err = workflows.Update(ctx, inner, []model.WorkflowStep{{WorkflowRef: outer.ID}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "loop")
	assert.Error(t, workflows.Update(ctx, inner, []model.WorkflowStep{{WorkflowRef: inner.ID}}))
	got, _ = workflows.Steps(ctx, inner.ID)
	assert.Equal(t, "build", got[0].Job) // unchanged

	// Reaches says which workflows run one, outer among them.
	users, err := workflows.UsedBy(ctx, inner.ID)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "outer", users[0].Name)

	err = workflows.Delete(ctx, inner.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "outer")
	assert.NoError(t, workflows.Delete(ctx, outer.ID))
	assert.NoError(t, workflows.Delete(ctx, inner.ID))
}

// A run's step copies the workflow a step names, and keeps the run it
// made of it.
func TestWorkflowRunStorage_WorkflowStep(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	workflows := storage.NewWorkflowStorage(db)
	runs := storage.NewWorkflowRunStorage(db)

	inner := &model.Workflow{Name: "inner"}
	assert.NoError(t, workflows.Create(ctx, inner, steps("infra", "build")))
	run := &model.WorkflowRun{WorkflowID: model.NewID(), Name: "outer"}
	assert.NoError(t, runs.Create(ctx, run, []model.WorkflowStep{{WorkflowRef: inner.ID}}))
	rows, err := runs.Steps(ctx, run.ID)
	assert.NoError(t, err)
	assert.Equal(t, inner.ID, rows[0].WorkflowRef)

	rows[0].WorkflowRunRef, rows[0].State = model.NewID(), model.StateOK
	assert.NoError(t, runs.UpdateStep(ctx, &rows[0]))
	rows, _ = runs.Steps(ctx, run.ID)
	assert.True(t, rows[0].WorkflowRunRef != "")
	assert.Equal(t, model.StateOK, rows[0].State)
}

// A workflow reaches the workflows its steps run, through any depth, and
// itself; not the ones that run it.
func TestWorkflowStorage_Reaches(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	workflows := storage.NewWorkflowStorage(newDB(t))
	inner := &model.Workflow{Name: "inner"}
	assert.NoError(t, workflows.Create(ctx, inner, steps("infra", "build")))
	middle := &model.Workflow{Name: "middle"}
	assert.NoError(t, workflows.Create(ctx, middle, []model.WorkflowStep{{WorkflowRef: inner.ID}}))
	outer := &model.Workflow{Name: "outer"}
	assert.NoError(t, workflows.Create(ctx, outer, []model.WorkflowStep{{Folder: "infra", Job: "test"}, {WorkflowRef: middle.ID}}))

	for _, c := range []struct {
		from, target string
		want         bool
	}{
		{outer.ID, inner.ID, true},
		{outer.ID, middle.ID, true},
		{inner.ID, inner.ID, true},
		{inner.ID, outer.ID, false},
		{middle.ID, outer.ID, false},
	} {
		got, err := workflows.Reaches(ctx, c.from, c.target)
		assert.NoError(t, err)
		assert.Equal(t, c.want, got)
	}
}

// Expand flattens the workflows the steps run into their jobs, in order.
func TestWorkflowStorage_Expand(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	workflows := storage.NewWorkflowStorage(newDB(t))
	inner := &model.Workflow{Name: "inner"}
	assert.NoError(t, workflows.Create(ctx, inner, steps("infra", "build", "infra", "test")))
	middle := &model.Workflow{Name: "middle"}
	assert.NoError(t, workflows.Create(ctx, middle, []model.WorkflowStep{{WorkflowRef: inner.ID}, {Folder: "infra", Job: "push"}}))

	flat, err := workflows.Expand(ctx, []model.WorkflowStep{{Folder: "infra", Job: "lint"}, {WorkflowRef: middle.ID}})
	assert.NoError(t, err)
	jobs := []string{}
	for _, s := range flat {
		assert.False(t, s.IsWorkflow())
		jobs = append(jobs, s.Job)
	}
	assert.Equal(t, []string{"lint", "build", "test", "push"}, jobs)

	gone, err := workflows.Expand(ctx, []model.WorkflowStep{{WorkflowRef: model.NewID()}})
	assert.NoError(t, err)
	assert.Len(t, gone, 0)
}

func TestWorkflowRunStorage_ListByWorkflow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewWorkflowRunStorage(newDB(t))
	base := time.Now().Add(-time.Hour).UTC().Truncate(time.Millisecond)
	var made []string // oldest first
	for i := range 5 {
		run := &model.WorkflowRun{ID: model.NewIDAt(base.Add(time.Duration(i) * time.Minute)), WorkflowID: "wf", Name: "deploy"}
		assert.NoError(t, runs.Create(ctx, run, steps("a", "build")))
		// Every run started at the same moment: the ID orders them.
		run.StartedAt = &base
		assert.NoError(t, runs.Update(ctx, run))
		made = append(made, run.ID)
	}
	assert.NoError(t, runs.Create(ctx, &model.WorkflowRun{WorkflowID: "other"}, steps("a", "build")))

	runIDs := func(list []model.WorkflowRun) []string {
		out := []string{}
		for _, run := range list {
			out = append(out, run.ID)
		}
		return out
	}
	first, err := runs.ListByWorkflow(ctx, "wf", storage.WorkflowRunFilter{Limit: 2})
	assert.NoError(t, err)
	assert.Equal(t, []string{made[4], made[3]}, runIDs(first))
	next, err := runs.ListByWorkflow(ctx, "wf", storage.WorkflowRunFilter{Limit: 2, Before: made[3]})
	assert.NoError(t, err)
	assert.Equal(t, []string{made[2], made[1]}, runIDs(next))
	back, err := runs.ListByWorkflow(ctx, "wf", storage.WorkflowRunFilter{Limit: 2, After: made[2]})
	assert.NoError(t, err)
	assert.Equal(t, runIDs(first), runIDs(back))
	all, err := runs.ListByWorkflow(ctx, "wf", storage.WorkflowRunFilter{})
	assert.NoError(t, err)
	assert.Len(t, all, 5)
	none, err := runs.ListByWorkflow(ctx, "wf", storage.WorkflowRunFilter{Before: "nope"})
	assert.NoError(t, err)
	assert.Len(t, none, 0)
}
