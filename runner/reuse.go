package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// ErrReuse refuses a run under the ID of a run of another job or
// directory, or under an ID that is no run ID.
var ErrReuse = errors.New("the run is of another job")

// replace clears what a run under the ID a new run reuses left: its row,
// its recording, its summary, and its artifacts' records and directory.
// The run must be of the same job in the same directory, and the job's
// lock, which the caller holds, keeps it from starting; a session of
// the run that is still settling is in the way, with ErrBusy. A new ID,
// or one with no row, has nothing to clear.
func (m *Manager) replace(ctx context.Context, run *model.Run) (*Session, error) {
	if run.ID == "" {
		return nil, nil
	}
	if !model.ValidID(run.ID) {
		return nil, fmt.Errorf("%w: invalid run id %q", ErrReuse, run.ID)
	}
	if session, ok := m.Get(run.ID); ok {
		select {
		case <-session.Done():
		default:
			return session, ErrBusy
		}
	}
	old, err := m.opts.Runs.Get(ctx, run.ID)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if old.Key() != run.Key() {
		return nil, ErrReuse
	}
	if m.opts.Recordings != nil {
		if path, err := m.opts.Recordings.Path(old); err == nil {
			_ = os.Remove(path)
		}
		if old.CastPath != "" {
			_ = os.Remove(filepath.Join(m.opts.Recordings.Root, filepath.FromSlash(old.CastPath)))
		}
	}
	if m.opts.Summaries != nil {
		if err := m.opts.Summaries.Delete(ctx, old.ID); err != nil {
			return nil, err
		}
	}
	if m.opts.Artifacts != nil {
		if err := m.opts.Artifacts.DeleteByRun(ctx, old.ID); err != nil {
			return nil, err
		}
	}
	if dir := m.artifactDir(old); dir != "" {
		_ = os.RemoveAll(dir)
	}
	return nil, m.opts.Runs.Delete(ctx, old.ID)
}
