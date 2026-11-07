//go:build integration

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
	assert.ErrorContains(t, workflows.Create(ctx, &model.Workflow{Name: "off"}, off), "a workflow needs an enabled step")

	// A prompt or a summary step runs a workflow with every job step
	// disabled; the kinds, the description and the step toggles round
	// trip.
	asks := &model.Workflow{Name: "asks", Description: "summarise the build", UserID: "u-ops", RunAsUserID: "u-miguel"}
	summary := model.WorkflowStep{Kind: model.StepSummary, Folder: "malta", Dir: "dockerfiles", Job: model.PromptJob}
	assert.NoError(t, workflows.Create(ctx, asks, append(off, model.WorkflowStep{Folder: "a", Job: "c"}, summary)))
	read, err := workflows.Get(ctx, asks.ID)
	assert.NoError(t, err)
	assert.Equal(t, "summarise the build", read.Description)
	assert.Equal(t, "u-miguel", read.RunAsUserID, "the run-as user round trips")
	toggled, err := workflows.Steps(ctx, asks.ID)
	assert.NoError(t, err)
	assert.True(t, toggled[0].Disabled)
	assert.False(t, toggled[1].Disabled)
	assert.Equal(t, model.StepJob, toggled[1].Kind)
	assert.Equal(t, model.StepSummary, toggled[2].Kind)
	assert.Equal(t, "malta/dockerfiles", toggled[2].Place())
	read.RunAsUserID = ""
	assert.NoError(t, workflows.Update(ctx, read, toggled))
	read, err = workflows.Get(ctx, asks.ID)
	assert.NoError(t, err)
	assert.Equal(t, "", read.RunAsUserID, "cleared: whoever starts it")
	alone := &model.Workflow{Name: "alone"}
	assert.NoError(t, workflows.Create(ctx, alone, []model.WorkflowStep{{Folder: "_", Job: model.PromptJob, Prompt: "say ok"}}))
	one, err := workflows.Steps(ctx, alone.ID)
	assert.NoError(t, err)
	assert.Len(t, one, 1)
	assert.Equal(t, model.StepPrompt, one[0].Kind, "a row's kind from its fields")
	assert.Error(t, workflows.Create(ctx, &model.Workflow{}, steps("a", "b")), "no name")
	assert.ErrorContains(t, workflows.Create(ctx, &model.Workflow{Name: "half"}, steps("a", "")), "step 1: a job step names a folder and a job")
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
	step, err := runs.StepOf(ctx, "run1")
	assert.NoError(t, err)
	assert.Equal(t, run.ID, step.WorkflowRunID)
	_, err = runs.StepOf(ctx, "")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = runs.StepOf(ctx, "other")
	assert.ErrorIs(t, err, storage.ErrNotFound)

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

// A step runs another workflow. A workflow may not
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

	assert.Error(t, workflows.Create(ctx, &model.Workflow{Name: "both"}, []model.WorkflowStep{{Folder: "infra", Job: "build", WorkflowRef: inner.ID}}), "a step names a job or a workflow")
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

// A prompt step keeps its prompt and its place, the root or a directory
// of a folder; a run's step copies the prompt and keeps Claude's answer.
func TestWorkflowStorage_PromptStep(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	workflows := storage.NewWorkflowStorage(db)
	runs := storage.NewWorkflowRunStorage(db)

	ask := model.WorkflowStep{Folder: "malta", Dir: "docs", Job: model.PromptJob, Prompt: "fetch the weather"}
	wf := &model.Workflow{Name: "weather"}
	assert.NoError(t, workflows.Create(ctx, wf, append(steps("infra", "build"), ask)))
	got, err := workflows.Steps(ctx, wf.ID)
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.True(t, got[1].IsPrompt())
	assert.Equal(t, "malta/docs", got[1].Place())

	bad := []model.WorkflowStep{
		{Folder: "_", Job: "build", Prompt: "which job"},
		{Folder: "_", Job: model.PromptJob, WorkflowRef: wf.ID, Prompt: "both"},
		{Kind: model.StepPrompt, Folder: "_", Job: model.PromptJob, Prompt: "  "},
		{Kind: model.StepSummary, Folder: "_", Job: model.PromptJob, Prompt: "a summary takes none"},
		{Kind: "other", Folder: "a", Job: "b"},
	}
	for _, step := range bad {
		assert.Error(t, workflows.Create(ctx, &model.Workflow{Name: "bad"}, []model.WorkflowStep{step}))
	}

	// A step with no folder asks at the workspaces root, with no workspace.
	sweep := &model.Workflow{Name: "sweep"}
	assert.NoError(t, workflows.Create(ctx, sweep, []model.WorkflowStep{{Kind: model.StepSummary, Job: model.PromptJob}}))
	noneSteps, err := workflows.Steps(ctx, sweep.ID)
	assert.NoError(t, err)
	assert.Equal(t, "", noneSteps[0].Folder)
	assert.Equal(t, "", noneSteps[0].Place())

	run := &model.WorkflowRun{WorkflowID: wf.ID, Name: wf.Name}
	assert.NoError(t, runs.Create(ctx, run, got))
	rows, err := runs.Steps(ctx, run.ID)
	assert.NoError(t, err)
	assert.Equal(t, "fetch the weather", rows[1].Prompt)
	assert.Equal(t, model.StepPrompt, rows[1].Kind, "a run's step keeps its kind")
	rows[1].Answer, rows[1].State = "sunny, 21 °C", model.StateOK
	assert.NoError(t, runs.UpdateStep(ctx, &rows[1]))
	rows, _ = runs.Steps(ctx, run.ID)
	assert.Equal(t, "sunny, 21 °C", rows[1].Answer)
}

// Answered lists a workflow's finished runs with an answer, newest
// first, without the run named and the running ones; Number counts a
// run's place among its workflow's runs.
func TestWorkflowRunStorage_Answered(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewWorkflowRunStorage(newDB(t))
	flow := model.NewID()
	made := []*model.WorkflowRun{}
	for i, answer := range []string{"one", "", "three", "four"} {
		run := &model.WorkflowRun{WorkflowID: flow, Name: "w"}
		assert.NoError(t, runs.Create(ctx, run, nil))
		run.State, run.Answer = model.StateOK, answer
		if i == 3 {
			run.State = model.StateRunning
		}
		assert.NoError(t, runs.Update(ctx, run))
		made = append(made, run)
	}
	got, err := runs.Answered(ctx, flow, made[2].ID, 3)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "one", got[0].Answer)
	got, _ = runs.Answered(ctx, flow, "", 3)
	assert.Len(t, got, 2)
	assert.Equal(t, "three", got[0].Answer)

	n, err := runs.Number(ctx, made[2])
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
}

func TestWorkflowRunStorage_Number(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewWorkflowRunStorage(newDB(t))
	run := &model.WorkflowRun{WorkflowID: model.NewID(), Name: "w"}
	assert.NoError(t, runs.Create(ctx, run, nil))
	n, err := runs.Number(ctx, run)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
}

// History lists every workflow run; Delete removes one
// with its steps and leaves the others and the workflow.
func TestWorkflowRunStorage_HistoryDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewWorkflowRunStorage(newDB(t))
	first := &model.WorkflowRun{WorkflowID: "wf", Name: "deploy"}
	assert.NoError(t, runs.Create(ctx, first, steps("a", "build")))
	second := &model.WorkflowRun{WorkflowID: "wf", Name: "deploy"}
	assert.NoError(t, runs.Create(ctx, second, steps("a", "build", "a", "push")))
	history, err := runs.History(ctx)
	assert.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Contains(t, []string{history[0].ID, history[1].ID}, first.ID)
	assert.Contains(t, []string{history[0].ID, history[1].ID}, second.ID)

	assert.NoError(t, runs.Delete(ctx, first.ID))
	_, err = runs.Get(ctx, first.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	chain, err := runs.Steps(ctx, first.ID)
	assert.NoError(t, err)
	assert.Len(t, chain, 0)
	chain, _ = runs.Steps(ctx, second.ID)
	assert.Len(t, chain, 2)
	history, _ = runs.History(ctx)
	assert.Len(t, history, 1)
}

// Every method of the workflows returns the driver's error.
func TestWorkflowStorage_Closed(t *testing.T) {
	t.Parallel()
	flows := storage.NewWorkflowStorage(closedDB(t))
	ctx := context.Background()
	workflow := &model.Workflow{Name: "deploy"}
	steps := []model.WorkflowStep{{Folder: "acme", Job: "build"}}

	assert.ErrorContains(t, flows.Create(ctx, workflow, steps), errClosed)
	_, err := flows.Get(ctx, "w1")
	assert.ErrorContains(t, err, errClosed)
	_, err = flows.Steps(ctx, "w1")
	assert.ErrorContains(t, err, errClosed)
	_, err = flows.List(ctx)
	assert.ErrorContains(t, err, errClosed)
	_, err = flows.UsedBy(ctx, "w1")
	assert.ErrorContains(t, err, errClosed)
	_, err = flows.Reaches(ctx, "w1", "w2")
	assert.ErrorContains(t, err, errClosed)
	_, err = flows.Expand(ctx, []model.WorkflowStep{{WorkflowRef: "w2"}})
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, flows.Update(ctx, workflow, steps), errClosed)
	assert.ErrorContains(t, flows.Delete(ctx, "w1"), errClosed)
}

// Every method of the workflow runs returns the driver's error.
func TestWorkflowRunStorage_Closed(t *testing.T) {
	t.Parallel()
	runs := storage.NewWorkflowRunStorage(closedDB(t))
	ctx := context.Background()
	run := &model.WorkflowRun{WorkflowID: "w1"}

	assert.ErrorContains(t, runs.Create(ctx, run, []model.WorkflowStep{{Folder: "acme", Job: "build"}}), errClosed)
	_, err := runs.Get(ctx, "r1")
	assert.ErrorContains(t, err, errClosed)
	_, err = runs.Steps(ctx, "r1")
	assert.ErrorContains(t, err, errClosed)
	_, err = runs.List(ctx, 0)
	assert.ErrorContains(t, err, errClosed)
	_, err = runs.History(ctx)
	assert.ErrorContains(t, err, errClosed)
	_, err = runs.Answered(ctx, "w1", "", 5)
	assert.ErrorContains(t, err, errClosed)
	_, err = runs.Number(ctx, run)
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, runs.Update(ctx, run), errClosed)
	assert.ErrorContains(t, runs.UpdateStep(ctx, &model.WorkflowRunStep{ID: "s1"}), errClosed)
	assert.ErrorContains(t, runs.Delete(ctx, "r1"), errClosed)
	assert.ErrorContains(t, runs.Interrupt(ctx, "lost"), errClosed)
}

// The statements inside a transaction hand their error back: the tables
// a workflow and a workflow run are written across, one gone at a time.
func TestWorkflowStorage_Broken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	workflow := &model.Workflow{ID: "w1", Name: "deploy"}
	steps := []model.WorkflowStep{{Folder: "acme", Job: "build"}}

	for _, table := range []string{model.WorkflowTable, model.WorkflowStepTable} {
		flows := storage.NewWorkflowStorage(brokenDB(t, table))
		assert.ErrorContains(t, flows.Create(ctx, &model.Workflow{Name: "deploy"}, steps), errNoTable, table)
		assert.ErrorContains(t, flows.Update(ctx, workflow, steps), errNoTable, table)
		assert.ErrorContains(t, flows.Delete(ctx, "w1"), errNoTable, table)
	}

	for _, table := range []string{model.WorkflowRunTable, model.WorkflowRunStepTable} {
		runs := storage.NewWorkflowRunStorage(brokenDB(t, table))
		assert.ErrorContains(t, runs.Create(ctx, &model.WorkflowRun{WorkflowID: "w1"}, steps), errNoTable, table)
		assert.ErrorContains(t, runs.Delete(ctx, "r1"), errNoTable, table)
		assert.ErrorContains(t, runs.Interrupt(ctx, "lost"), errNoTable, table)
	}
}

// A step that runs a workflow is checked against the workflows there
// are, and the check reads them: a workflow that is not there, one that
// runs this one, and the reads that answer neither.
func TestWorkflowStorage_RefErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	flows := storage.NewWorkflowStorage(newDB(t))

	inner := &model.Workflow{Name: "inner"}
	assert.NoError(t, flows.Create(ctx, inner, []model.WorkflowStep{{Folder: "acme", Job: "build"}}))
	outer := &model.Workflow{Name: "outer"}
	assert.NoError(t, flows.Create(ctx, outer, []model.WorkflowStep{{WorkflowRef: inner.ID}}))

	assert.ErrorContains(t, flows.Update(ctx, inner, []model.WorkflowStep{{WorkflowRef: inner.ID}}), "may not run itself")
	assert.ErrorContains(t, flows.Update(ctx, inner, []model.WorkflowStep{{WorkflowRef: "nope"}}), "no workflow nope")
	assert.ErrorContains(t, flows.Update(ctx, inner, []model.WorkflowStep{{WorkflowRef: outer.ID}}), "runs this workflow: a loop")

	// The read the loop check makes is the steps of the workflow a step
	// runs; without that table it answers neither yes nor no.
	gone := storage.NewWorkflowStorage(brokenDB(t, model.WorkflowStepTable))
	_, err := gone.Reaches(ctx, "w1", "w2")
	assert.ErrorContains(t, err, errNoTable)
	_, err = gone.Expand(ctx, []model.WorkflowStep{{WorkflowRef: "w2"}})
	assert.ErrorContains(t, err, errNoTable)

	// The workflow a step runs is there and the loop check cannot read
	// its steps: the check returns the read's error and decides nothing.
	losing := newDB(t)
	losingFlows := storage.NewWorkflowStorage(losing)
	kept := &model.Workflow{Name: "kept"}
	assert.NoError(t, losingFlows.Create(ctx, kept, []model.WorkflowStep{{Folder: "acme", Job: "build"}}))
	caller := &model.Workflow{Name: "caller"}
	assert.NoError(t, losingFlows.Create(ctx, caller, []model.WorkflowStep{{Folder: "acme", Job: "build"}}))
	drop(t, losing, model.WorkflowStepTable)
	assert.ErrorContains(t, losingFlows.Update(ctx, caller, []model.WorkflowStep{{WorkflowRef: kept.ID}}), errNoTable)
}

// A workflow two steps run is expanded once: the second step is the
// same workflow and adds nothing.
func TestWorkflowStorage_ExpandTwice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	flows := storage.NewWorkflowStorage(newDB(t))

	inner := &model.Workflow{Name: "inner"}
	assert.NoError(t, flows.Create(ctx, inner, []model.WorkflowStep{{Folder: "acme", Job: "build"}}))

	flat, err := flows.Expand(ctx, []model.WorkflowStep{{WorkflowRef: inner.ID}, {WorkflowRef: inner.ID}})
	assert.NoError(t, err)
	assert.Len(t, flat, 1)
	assert.Equal(t, "build", flat[0].Job)
}

// The check of a step's workflow reads the row, and a read that answers
// neither the row nor "not there" is its own error.
func TestWorkflowStorage_RefReadFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	flows := storage.NewWorkflowStorage(db)
	inner := &model.Workflow{Name: "inner"}
	assert.NoError(t, flows.Create(ctx, inner, []model.WorkflowStep{{Folder: "acme", Job: "build"}}))
	outer := &model.Workflow{Name: "outer"}
	assert.NoError(t, flows.Create(ctx, outer, []model.WorkflowStep{{Folder: "acme", Job: "build"}}))

	widen(t, db, model.WorkflowTable)
	assert.ErrorContains(t, flows.Update(ctx, outer, []model.WorkflowStep{{WorkflowRef: inner.ID}}), errNoField)
}

// A delete clears the steps and then the workflow, and stops at the one
// that failed. An interrupt writes the steps that ran, the steps that
// waited and then the run, and stops at the one that failed.
func TestWorkflowStorage_WriteErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	dropping := newDB(t)
	flows := storage.NewWorkflowStorage(dropping)
	workflow := &model.Workflow{Name: "deploy"}
	assert.NoError(t, flows.Create(ctx, workflow, []model.WorkflowStep{{Folder: "acme", Job: "build"}}))
	refuse(t, dropping, model.WorkflowTable, "DELETE")
	assert.ErrorContains(t, flows.Delete(ctx, workflow.ID), errRefused)

	steps := newDB(t)
	withSteps := storage.NewWorkflowStorage(steps)
	stepped := &model.Workflow{Name: "deploy"}
	assert.NoError(t, withSteps.Create(ctx, stepped, []model.WorkflowStep{{Folder: "acme", Job: "build"}}))
	refuse(t, steps, model.WorkflowStepTable, "DELETE")
	assert.ErrorContains(t, withSteps.Delete(ctx, stepped.ID), errRefused)

	// A run whose step waits and none of whose steps runs: the first
	// statement of the interrupt writes nothing and the second writes
	// the step that waited.
	waiting := newDB(t)
	runs := storage.NewWorkflowRunStorage(waiting)
	assert.NoError(t, runs.Create(ctx, &model.WorkflowRun{WorkflowID: "w1"}, []model.WorkflowStep{{Folder: "acme", Job: "build"}}))
	refuse(t, waiting, model.WorkflowRunStepTable, "UPDATE")
	assert.ErrorContains(t, runs.Interrupt(ctx, "lost"), errRefused)

	// The same run with its steps left alone: the interrupt reaches the
	// run's own row.
	lost := newDB(t)
	byRun := storage.NewWorkflowRunStorage(lost)
	assert.NoError(t, byRun.Create(ctx, &model.WorkflowRun{WorkflowID: "w1"}, []model.WorkflowStep{{Folder: "acme", Job: "build"}}))
	refuse(t, lost, model.WorkflowRunTable, "UPDATE")
	assert.ErrorContains(t, byRun.Interrupt(ctx, "lost"), errRefused)
}

// A workflow another workflow runs is kept, named by what runs it.
func TestWorkflowStorage_DeleteInUse(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	flows := storage.NewWorkflowStorage(newDB(t))

	inner := &model.Workflow{Name: "inner"}
	assert.NoError(t, flows.Create(ctx, inner, []model.WorkflowStep{{Folder: "acme", Job: "build"}}))
	outer := &model.Workflow{Name: "outer"}
	assert.NoError(t, flows.Create(ctx, outer, []model.WorkflowStep{{WorkflowRef: inner.ID}}))

	err := flows.Delete(ctx, inner.ID)
	assert.ErrorIs(t, err, storage.ErrInUse)
	assert.ErrorContains(t, err, "outer runs it")
}
