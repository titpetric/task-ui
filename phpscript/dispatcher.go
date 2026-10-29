package phpscript

import (
	"context"
)

// Dispatcher starts a Claude run for a module. `Task::dispatch` reaches it
// with the comment a user wrote and the path of a markdown file under the
// module's root; the prompt is that file, then the comment, then the run's own
// environment, the artifact tree among it.
//
// It returns the identifier of what it started, which the module records
// beside the comment, so a page links the run that answered it.
type Dispatcher interface {
	Dispatch(ctx context.Context, m *Module, comment, action string) (string, error)
}
