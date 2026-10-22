package api

// WorkspaceRequest adds a workspace from a git repository, or changes
// one. A workspace's folder is the name its projects clone under; the
// URL, the folder and the branch hold on create only, an update writes
// the title and the description, and the sort and Recursive when the
// request has them.
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
}
