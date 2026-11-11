package phpscript

import (
	"context"
)

// Runs answers which run a module's dispatch turned into. `Task::run`
// reaches it with the note identifier a dispatch returned.
//
// A dispatch opens a note and starts the run behind it, in that order,
// so the run does not exist yet when the dispatch answers. The module
// records the note and asks for the run when it draws a page, which is
// later and by then there is one.
//
// It answers "" for a note with no run yet, and for no note at all. A
// page draws a label either way; an empty one is a run that has not
// started.
type Runs interface {
	RunOf(ctx context.Context, note string) string
}
