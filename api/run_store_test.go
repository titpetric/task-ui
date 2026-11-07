package api_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// memRuns are runs in memory, newest last.
type memRuns struct{ runs []model.Run }

// Get returns a copy of the run with an ID, storage.ErrNotFound for none.
func (m *memRuns) Get(_ context.Context, id string) (*model.Run, error) {
	for _, run := range m.runs {
		if run.ID == id {
			return &run, nil
		}
	}
	return nil, storage.ErrNotFound
}

// List returns the runs of a filter's folder and job, newest first.
func (m *memRuns) List(_ context.Context, filter storage.RunFilter) ([]model.Run, error) {
	out := []model.Run{}
	for i := len(m.runs) - 1; i >= 0; i-- {
		if matches(m.runs[i], filter) {
			out = append(out, m.runs[i])
		}
	}
	return out, nil
}

// Last returns the newest run of a filter's folder and job, storage.ErrNotFound for none.
func (m *memRuns) Last(ctx context.Context, filter storage.RunFilter) (*model.Run, error) {
	runs, err := m.List(ctx, filter)
	if err != nil || len(runs) == 0 {
		return nil, storage.ErrNotFound
	}
	return &runs[0], nil
}

// Active returns the runs still running, newest first.
func (m *memRuns) Active(ctx context.Context) ([]model.Run, error) {
	return m.List(ctx, storage.RunFilter{State: model.StateRunning})
}

// matches reports whether a run answers the fields of a filter the fake reads.
func matches(run model.Run, filter storage.RunFilter) bool {
	switch {
	case filter.Folder != "" && filter.Folder != run.Folder:
		return false
	case filter.Job != "" && filter.Job != run.Job:
		return false
	case filter.State != "" && filter.State != run.State:
		return false
	}
	return true
}

// A store of runs is what the handlers take: one by ID, a folder's
// listing, the newest run of a job and the ones still going.
func TestRunStore(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var store api.RunStore = &memRuns{runs: []model.Run{
		{ID: "r1", Folder: "infra", Job: "plan", State: model.StateOK},
		{ID: "r2", Folder: "infra", Job: "plan", State: model.StateFailed},
		{ID: "r3", Folder: "malta", Job: "build", State: model.StateRunning},
	}}

	run, err := store.Get(ctx, "r1")
	assert.NoError(t, err)
	assert.Equal(t, "plan", run.Job)
	_, err = store.Get(ctx, "gone")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	runs, err := store.List(ctx, storage.RunFilter{Folder: "infra"})
	assert.NoError(t, err)
	assert.Len(t, runs, 2)
	last, err := store.Last(ctx, storage.RunFilter{Folder: "infra", Job: "plan"})
	assert.NoError(t, err)
	assert.Equal(t, "r2", last.ID)
	_, err = store.Last(ctx, storage.RunFilter{Folder: "gozo"})
	assert.ErrorIs(t, err, storage.ErrNotFound)

	active, err := store.Active(ctx)
	assert.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Equal(t, "r3", active[0].ID)
}
