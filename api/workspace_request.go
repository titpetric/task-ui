package api

// WorkspaceRequest adds a workspace, a directory under the root, or
// changes one. A workspace's folder is the name its projects clone
// under. On create, a URL clones the home repository into the
// workspace's home, on Branch; the folder holds on create only. An
// update writes the title and the description, and the sort and
// Recursive when the request has them; the home changes through
// WorkspaceHomeRequest.
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

// WorkspaceHomeRequest replaces the home of a workspace with a clone of
// a repository, on Branch when named.
type WorkspaceHomeRequest struct {
	URL    string `json:"url"`
	Branch string `json:"branch"`
}
