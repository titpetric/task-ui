package app

import (
	"context"

	"github.com/titpetric/task-ui/storage"
)

// ModuleRuns answers which run a module's dispatch turned into. It is
// what `Task::run($note)` reaches.
//
// A dispatch is an ordinary prompt: the note is the thread and the run
// is what the board started on it, so the note holds the run once there
// is one. The module keeps the note beside its own row and asks here
// when it draws the page.
type ModuleRuns struct {
	Notes *storage.NoteStorage
}

// NewModuleRuns returns the lookup over the instance's notes.
func NewModuleRuns(notes *storage.NoteStorage) *ModuleRuns {
	return &ModuleRuns{Notes: notes}
}

// RunOf is the run of a note, or "" where the note is not there, holds
// no run yet, or the instance keeps no notes. A label with no run is a
// run that has not started.
func (r *ModuleRuns) RunOf(ctx context.Context, note string) string {
	if r == nil || r.Notes == nil {
		return ""
	}
	row, err := r.Notes.Get(ctx, note)
	if err != nil || row == nil {
		return ""
	}
	return row.RunID
}
