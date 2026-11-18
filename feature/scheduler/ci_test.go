package scheduler_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/scheduler"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

type fakeCISchedules struct {
	rows    []model.Schedule
	created []model.Schedule
	updated []model.Schedule
}

func (f *fakeCISchedules) List(context.Context) ([]model.Schedule, error) { return f.rows, nil }

func (f *fakeCISchedules) Create(_ context.Context, schedule *model.Schedule) error {
	schedule.ID = "new"
	f.rows = append(f.rows, *schedule)
	f.created = append(f.created, *schedule)
	return nil
}

func (f *fakeCISchedules) Update(_ context.Context, schedule *model.Schedule) error {
	f.updated = append(f.updated, *schedule)
	return nil
}

type fakeCISearch struct{ rows []storage.SearchDocument }

func (f *fakeCISearch) Targets(context.Context, string) ([]storage.SearchDocument, error) {
	return f.rows, nil
}

// A ci:daily target becomes one disabled @daily row; a second run adds none,
// and the row whose target left is disabled.
func TestCIReconciler(t *testing.T) {
	t.Parallel()
	var _ scheduler.CISchedules = (*fakeCISchedules)(nil)
	var _ scheduler.CISearch = (*fakeCISearch)(nil)
	ctx := context.Background()
	schedules := &fakeCISchedules{}
	search := &fakeCISearch{rows: []storage.SearchDocument{
		{Kind: storage.SearchJobKind, FolderID: "acme", Target: "ci:daily", Name: "ci:daily"},
	}}
	reconciler := &scheduler.CIReconciler{Schedules: schedules, Search: search}

	added, err := reconciler.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, added)
	assert.Len(t, schedules.created, 1)
	assert.Equal(t, scheduler.CIExpression, schedules.created[0].Expression)
	assert.Equal(t, int64(0), schedules.created[0].Enabled)
	assert.Equal(t, scheduler.CISource, schedules.created[0].Source)

	added, err = reconciler.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, added)

	// An admin's own schedule for the same job is left alone.
	schedules.rows = append(schedules.rows, model.Schedule{Folder: "other", Job: "ci:daily", Expression: "0 3 * * *", Enabled: 1})
	search.rows = append(search.rows, storage.SearchDocument{FolderID: "other", Target: "ci:daily"})
	added, err = reconciler.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, added)

	// The target leaving disables the managed row and leaves the user's row.
	schedules.rows[0].Enabled = 1
	search.rows = nil
	_, err = reconciler.Run(ctx)
	assert.NoError(t, err)
	assert.Len(t, schedules.updated, 1)
	assert.Equal(t, "acme", schedules.updated[0].Folder)
	assert.Equal(t, int64(0), schedules.updated[0].Enabled)
	assert.Equal(t, scheduler.CISource, schedules.updated[0].Source)
}

// A target in a directory has no schedule row to name it, and is skipped.
func TestCIReconciler_Nested(t *testing.T) {
	t.Parallel()
	schedules := &fakeCISchedules{}
	search := &fakeCISearch{rows: []storage.SearchDocument{{FolderID: "acme", Target: "ci:daily", Dir: "sub"}}}
	added, err := (&scheduler.CIReconciler{Schedules: schedules, Search: search}).Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, added)
}
