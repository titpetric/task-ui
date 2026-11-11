package api

import (
	"strings"

	"github.com/titpetric/task-ui/hours"
)

// WorkspaceRequest adds a workspace, a directory under the root, or
// changes one. A workspace's folder is the name its projects clone
// under. On create, a URL clones the home repository into the
// workspace's home, on Branch; the folder holds on create only. An
// update writes the title and the description, and the sort, Recursive,
// the sidebar limit and the hours when the request has them; the home changes
// through WorkspaceHomeRequest.
type WorkspaceRequest struct {
	URL         string `json:"url"`
	Folder      string `json:"folder"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Branch      string `json:"branch"`
	// Sort orders the workspaces, lowest first; on an update, left out
	// keeps it.
	Sort *int64 `json:"sort,omitempty"`
	// Recursive lists the folders inside the workspace as deep as they
	// go, the default; false lists the workspace and its top-level
	// folders alone. On an update, left out keeps it.
	Recursive *bool `json:"recursive,omitempty"`
	// SidebarLimit is how many of the workspace's projects the sidebar
	// lists, the most recently committed first; 0 lists every one, by
	// name. On an update, left out keeps it; a create takes none.
	SidebarLimit *int64 `json:"sidebar_limit,omitempty"`
	// Hours says when the runs of the workspace's folders may start, as
	// the hours package reads them, one rule per line; empty allows any
	// time. Left out, they stay unchanged.
	Hours *string `json:"hours,omitempty"`
}

// hours returns the request's hours, trimmed, and whether it has them:
// an error for hours the hours package does not read.
func (req WorkspaceRequest) hours() (string, bool, error) {
	if req.Hours == nil {
		return "", false, nil
	}
	text := strings.TrimSpace(*req.Hours)
	if _, err := hours.Parse(text); err != nil {
		return "", true, err
	}
	return text, true, nil
}

// WorkspaceHomeRequest replaces the home of a workspace with a clone of
// a repository, on Branch when named.
type WorkspaceHomeRequest struct {
	URL    string `json:"url"`
	Branch string `json:"branch"`
}
