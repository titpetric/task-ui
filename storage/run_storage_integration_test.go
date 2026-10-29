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

func TestRunStorage_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))

	run := &model.Run{Folder: "infra", Job: "deploy", UserName: "ops", TermCols: 100, TermRows: 30}
	assert.NoError(t, runs.Create(ctx, run))
	assert.True(t, model.ValidID(run.ID))
	assert.Equal(t, model.StateRunning, run.State)
	assert.NotNil(t, run.StartedAt)
	assert.Nil(t, run.FinishedAt)

	assert.Error(t, runs.Create(ctx, &model.Run{Folder: "infra"}), "no job")
}

func TestRunStorage_Get(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	run := &model.Run{Folder: "infra", Job: "deploy", Forced: 1}
	assert.NoError(t, runs.Create(ctx, run))

	got, err := runs.Get(ctx, run.ID)
	assert.NoError(t, err)
	assert.Equal(t, "infra/deploy", got.Key())
	assert.Equal(t, int64(1), got.Forced)
	assert.NotNil(t, got.StartedAt)

	_, err = runs.Get(ctx, "missing")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestRunStorage_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	base := time.Now().Add(-time.Hour)
	for i, job := range []string{"deploy", "plan", "deploy"} {
		started := base.Add(time.Duration(i) * time.Minute)
		assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "infra", Job: job, UserID: "u1", StartedAt: &started}))
	}
	assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "web", Job: "build", UserID: "u2", UserName: "ops"}))

	all, err := runs.List(ctx, storage.RunFilter{})
	assert.NoError(t, err)
	assert.Len(t, all, 4)
	assert.Equal(t, "web", all[0].Folder, "newest first")

	deploys, err := runs.List(ctx, storage.RunFilter{Folder: "infra", Job: "deploy"})
	assert.NoError(t, err)
	assert.Len(t, deploys, 2)

	limited, err := runs.List(ctx, storage.RunFilter{Limit: 1, UserID: "u1"})
	assert.NoError(t, err)
	assert.Len(t, limited, 1)
	assert.Equal(t, "u1", limited[0].UserID)

	byName, err := runs.List(ctx, storage.RunFilter{UserName: "ops"})
	assert.NoError(t, err)
	assert.Len(t, byName, 1)
	assert.Equal(t, "web", byName[0].Folder)

	none, err := runs.List(ctx, storage.RunFilter{State: model.StateOK})
	assert.NoError(t, err)
	assert.Len(t, none, 0)
}

func TestRunStorage_Active(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	running := &model.Run{Folder: "infra", Job: "deploy"}
	assert.NoError(t, runs.Create(ctx, running))
	done := &model.Run{Folder: "infra", Job: "plan"}
	assert.NoError(t, runs.Create(ctx, done))
	done.State = model.StateOK
	assert.NoError(t, runs.Finish(ctx, done))

	active, err := runs.Active(ctx)
	assert.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Equal(t, running.ID, active[0].ID)
}

func TestRunStorage_Last(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	_, err := runs.Last(ctx, storage.RunFilter{Folder: "infra", Job: "deploy"})
	assert.ErrorIs(t, err, storage.ErrNotFound)

	earlier := time.Now().Add(-time.Hour)
	assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "infra", Job: "deploy", UserName: "old", StartedAt: &earlier}))
	assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "infra", Job: "deploy", UserName: "new"}))
	last, err := runs.Last(ctx, storage.RunFilter{Folder: "infra", Job: "deploy"})
	assert.NoError(t, err)
	assert.Equal(t, "new", last.UserName)
}

func TestRunStorage_Latest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	base := time.Now().Add(-time.Hour)
	for i, run := range []model.Run{
		{Folder: "infra", Job: "deploy", State: model.StateFailed},
		{Folder: "infra", Job: "deploy", State: model.StateOK},
		{Folder: "infra", Job: "deploy"},
		{Folder: "infra", Job: "plan", State: model.StateKilled},
		{Folder: "web", Job: "build", State: model.StateOK},
	} {
		started := base.Add(time.Duration(i) * time.Minute)
		run.StartedAt = &started
		assert.NoError(t, runs.Create(ctx, &run))
	}

	latest, err := runs.Latest(ctx, []string{"infra"})
	assert.NoError(t, err)
	assert.Len(t, latest, 2)
	assert.Equal(t, "infra/deploy", latest[0].Key())
	assert.Equal(t, model.StateOK, latest[0].State, "the running one is not finished")
	assert.Equal(t, "infra/plan", latest[1].Key())
	assert.Equal(t, model.StateKilled, latest[1].State)

	all, err := runs.Latest(ctx, nil)
	assert.NoError(t, err)
	assert.Len(t, all, 3)
	assert.Equal(t, "web/build", all[2].Key())
}

func TestRunStorage_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	run := &model.Run{Folder: "infra", Job: "deploy"}
	assert.NoError(t, runs.Create(ctx, run))
	run.CastPath = "infra/deploy/x.cast"
	assert.NoError(t, runs.Update(ctx, run))
	got, err := runs.Get(ctx, run.ID)
	assert.NoError(t, err)
	assert.Equal(t, "infra/deploy/x.cast", got.CastPath)
}

func TestRunStorage_Finish(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	run := &model.Run{Folder: "infra", Job: "deploy"}
	assert.NoError(t, runs.Create(ctx, run))

	assert.Error(t, runs.Finish(ctx, run), "still running")
	run.State = model.StateFailed
	run.ExitCode = 2
	assert.NoError(t, runs.Finish(ctx, run))
	got, err := runs.Get(ctx, run.ID)
	assert.NoError(t, err)
	assert.True(t, got.Finished())
	assert.Equal(t, int64(2), got.ExitCode)
	assert.NotNil(t, got.FinishedAt)
}

func TestRunStorage_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	run := &model.Run{Folder: "infra", Job: "deploy"}
	assert.NoError(t, runs.Create(ctx, run))
	assert.NoError(t, runs.Delete(ctx, run.ID))
	_, err := runs.Get(ctx, run.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestRunStorage_KillActive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "infra", Job: "deploy"}))
	assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "infra", Job: "plan"}))

	killed, err := runs.KillActive(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), killed)
	active, err := runs.Active(ctx)
	assert.NoError(t, err)
	assert.Len(t, active, 0)
	list, err := runs.List(ctx, storage.RunFilter{State: model.StateKilled})
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.NotNil(t, list[0].FinishedAt)
}

// A run queued when the instance went down never started: the next
// start marks it failed, saying so; a running run stays for KillActive.
func TestRunStorage_FailQueued(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	queued := &model.Run{Folder: "_", Job: "@host:dev1", Target: "dev1", State: model.StateQueued}
	assert.NoError(t, runs.Create(ctx, queued))
	assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "_", Job: "@host:dev2", Target: "dev2"}))

	failed, err := runs.FailQueued(ctx, model.QueuedLost)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), failed)
	got, err := runs.Get(ctx, queued.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.StateFailed, got.State)
	assert.Equal(t, "the instance restarted before it ran", got.Message)
	assert.NotNil(t, got.FinishedAt)
	running, err := runs.List(ctx, storage.RunFilter{State: model.StateRunning})
	assert.NoError(t, err)
	assert.Len(t, running, 1)
}

// ids returns the IDs of runs, in order.
func ids(runs []model.Run) []string {
	out := make([]string, 0, len(runs))
	for _, run := range runs {
		out = append(out, run.ID)
	}
	return out
}

func TestRunStorage_ListPages(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))
	base := time.Now().Add(-time.Hour)
	var made []string // oldest first
	for i := range 7 {
		started := base.Add(time.Duration(i) * time.Minute)
		folder := "infra"
		if i%2 == 1 {
			folder = "web"
		}
		run := &model.Run{ID: model.NewIDAt(started), Folder: folder, Job: "deploy", StartedAt: &started}
		assert.NoError(t, runs.Create(ctx, run))
		made = append(made, run.ID)
	}

	// The first page, then the page before its last row: full, newest first.
	first, err := runs.List(ctx, storage.RunFilter{Limit: 3})
	assert.NoError(t, err)
	assert.Equal(t, []string{made[6], made[5], made[4]}, ids(first))
	second, err := runs.List(ctx, storage.RunFilter{Limit: 3, Before: first[2].ID})
	assert.NoError(t, err)
	assert.Equal(t, []string{made[3], made[2], made[1]}, ids(second))

	// Back from the second page's first row lands on the first page.
	back, err := runs.List(ctx, storage.RunFilter{Limit: 3, After: second[0].ID})
	assert.NoError(t, err)
	assert.Equal(t, ids(first), ids(back))

	// The folders narrow in the query, so a page stays full.
	web, err := runs.List(ctx, storage.RunFilter{Limit: 2, Folders: []string{"web"}})
	assert.NoError(t, err)
	assert.Equal(t, []string{made[5], made[3]}, ids(web))
	both, err := runs.List(ctx, storage.RunFilter{Folders: []string{"web", "infra"}})
	assert.NoError(t, err)
	assert.Len(t, both, 7)
	older, err := runs.List(ctx, storage.RunFilter{Folders: []string{"infra"}, Before: made[4]})
	assert.NoError(t, err)
	assert.Equal(t, []string{made[2], made[0]}, ids(older))

	// A cursor on a run that is not there pages nothing.
	missing, err := runs.List(ctx, storage.RunFilter{Before: model.NewID()})
	assert.NoError(t, err)
	assert.Len(t, missing, 0)
	missing, err = runs.List(ctx, storage.RunFilter{After: "nope"})
	assert.NoError(t, err)
	assert.Len(t, missing, 0)
}

// Recorded lists the runs with a recording, oldest first; Busy names the
// runs in progress and the steps of a workflow run in progress.
func TestRunStorage_Recorded(t *testing.T) {
	t.Parallel()
	db := newDB(t)
	runs := storage.NewRunStorage(db)
	ctx := context.Background()
	older := time.Now().Add(-time.Hour)
	first := &model.Run{Folder: "infra", Job: "plan", State: model.StateOK, CastPath: "a.cast", StartedAt: &older}
	second := &model.Run{Folder: "infra", Job: "plan", State: model.StateRunning, CastPath: "b.cast"}
	none := &model.Run{Folder: "infra", Job: "plan", State: model.StateOK}
	for _, run := range []*model.Run{second, first, none} {
		assert.NoError(t, runs.Create(ctx, run))
	}
	recorded, err := runs.Recorded(ctx)
	assert.NoError(t, err)
	assert.Len(t, recorded, 2)
	assert.Equal(t, first.ID, recorded[0].ID)
}

func TestRunStorage_Busy(t *testing.T) {
	t.Parallel()
	db := newDB(t)
	runs := storage.NewRunStorage(db)
	flows := storage.NewWorkflowRunStorage(db)
	ctx := context.Background()
	running := &model.Run{Folder: "infra", Job: "plan"}
	step := &model.Run{Folder: "infra", Job: "plan", State: model.StateOK}
	done := &model.Run{Folder: "infra", Job: "plan", State: model.StateOK}
	for _, run := range []*model.Run{running, step, done} {
		assert.NoError(t, runs.Create(ctx, run))
	}
	flow := &model.WorkflowRun{WorkflowID: "w1", Name: "ship"}
	assert.NoError(t, flows.Create(ctx, flow, []model.WorkflowStep{{Folder: "infra", Job: "plan"}}))
	steps, _ := flows.Steps(ctx, flow.ID)
	steps[0].RunID = step.ID
	assert.NoError(t, flows.UpdateStep(ctx, &steps[0]))

	busy, err := runs.Busy(ctx)
	assert.NoError(t, err)
	assert.Equal(t, map[string]bool{running.ID: true, step.ID: true}, busy)
}

// TestRunStorage_CreateNormalisesTheClock covers the ordering of runs on
// a host that is not on UTC.
//
// The query compares and orders started_at as text, so two zones in one
// column sort by their offset text: without the
// conversion a run started an hour ago in CEST sorts above one started
// now, and the list is not newest first.
func TestRunStorage_CreateNormalisesTheClock(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runs := storage.NewRunStorage(newDB(t))

	zone := time.FixedZone("CEST", 2*60*60)
	earlier := time.Now().In(zone).Add(-time.Hour)
	assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "infra", Job: "deploy", UserName: "old", StartedAt: &earlier}))
	assert.NoError(t, runs.Create(ctx, &model.Run{Folder: "infra", Job: "deploy", UserName: "new"}))

	// The row the caller handed a zoned time keeps its instant, in UTC.
	rows, err := runs.List(ctx, storage.RunFilter{})
	assert.NoError(t, err)
	assert.Len(t, rows, 2)
	assert.Equal(t, "new", rows[0].UserName, "newest first")
	assert.Equal(t, time.UTC, rows[1].StartedAt.Location())
	assert.Equal(t, earlier.UTC(), rows[1].StartedAt.UTC())

	last, err := runs.Last(ctx, storage.RunFilter{Folder: "infra", Job: "deploy"})
	assert.NoError(t, err)
	assert.Equal(t, "new", last.UserName)
}

// History lists every run, oldest first, recorded or not.
func TestRunStorage_History(t *testing.T) {
	t.Parallel()
	db := newDB(t)
	runs := storage.NewRunStorage(db)
	ctx := context.Background()
	older := time.Now().Add(-time.Hour)
	first := &model.Run{Folder: "infra", Job: "plan", State: model.StateOK, StartedAt: &older}
	second := &model.Run{Folder: "infra", Job: "plan", State: model.StateRunning, CastPath: "b.cast"}
	for _, run := range []*model.Run{second, first} {
		assert.NoError(t, runs.Create(ctx, run))
	}
	history, err := runs.History(ctx)
	assert.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Equal(t, first.ID, history[0].ID)
	assert.Equal(t, second.ID, history[1].ID)
}
