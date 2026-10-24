package view

import (
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/git"
)

// WorkspaceHome is the home of a workspace as its settings page shows
// it: the folder ID and directory of the home, the repository of its
// checkout, what git knows about the checkout, and the pull job of
// the home's git skill, for a user who may run it.
type WorkspaceHome struct {
	Folder string
	Dir    string
	URL    string
	Branch string
	// Checkout is nil for an empty home, or one that is no checkout.
	Checkout *git.Info
	Pull     *folder.Job
}

// Empty reports a home with no repository.
func (h *WorkspaceHome) Empty() bool {
	return h == nil || h.URL == ""
}
