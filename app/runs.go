package app

import (
	"context"

	"github.com/titpetric/task-ui/storage"
)

// ModuleRuns answers which run a module's dispatch turned into, behind
// `Task::run($note)`.
//
// A dispatch is an ordinary prompt: the note is the thread, and the note holds
// the run once the board starts one. The module keeps the note beside its own
// row and reads it here when it draws the page.
type ModuleRuns struct {
	Notes *storage.NoteStorage
}

// NewModuleRuns returns the lookup over the instance's notes.
func NewModuleRuns(notes *storage.NoteStorage) *ModuleRuns {
	return &ModuleRuns{Notes: notes}
}

// RunOf is the run of a note: "" for a missing note, for a run that has not
// started, and on an instance that keeps no notes.
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
