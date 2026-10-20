package api

// WorkspaceRequest adds a workspace from a git repository, or changes
// one. A workspace's folder is the name its projects clone under; the
// URL, the folder and the branch hold on create only, an update writes
// the title and the description.
type WorkspaceRequest struct {
	URL         string `json:"url"`
	Folder      string `json:"folder"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Branch      string `json:"branch"`
}
