package phpfront

import (
	"context"
)

// Dispatcher starts a Claude run for a module. `Task::dispatch` reaches it
// with the comment a user wrote and the path of a markdown file under the
// module's root; the prompt is that file, then the comment, then the run's own
// environment, the artifact tree among it.
//
// The root is the tree the action file is resolved inside, which is the
// module's own for a plugin and the front end's tree for a ported page: a
// dispatcher takes a directory and not a module, because a module is the
// only thing the two callers do not share.
//
// The folder is where in the artifact tree the run writes, and the only
// part of the tree the prompt names. A module that gives each of its
// threads a folder passes it here and the run gets that folder to itself;
// empty is the top of the tree, which every run shares.
//
// It returns the identifier of what it started, which the module records
// beside the comment, so a page links the run that answered it.
type Dispatcher interface {
	Dispatch(ctx context.Context, root, comment, action, folder string) (string, error)
}
