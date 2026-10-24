package api

// WorkspaceRequest adds a workspace, a directory under the root, or
// changes one. A workspace's folder is the name its projects clone
// under. On create, a URL clones a repository into the workspace as a
// project, Project its folder or one from the URL, on Branch, and makes
// it the home project; the folder holds on create only. An update
// writes the title and the description, and the sort, Recursive and
// the home project when the request has them.
type WorkspaceRequest struct {
	URL         string `json:"url"`
	Folder      string `json:"folder"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Branch      string `json:"branch"`
	// Project is the folder, inside the workspace, the URL clones into on
	// create; empty takes the repository's name.
	Project string `json:"project,omitempty"`
	// Sort orders the workspaces, lowest first; on an update, left out
	// keeps it.
	Sort *int64 `json:"sort,omitempty"`
	// Recursive lists the folders inside the workspace as deep as they
	// go, the default; false lists the workspace and its top-level
	// folders alone. On an update, left out keeps it.
	Recursive *bool `json:"recursive,omitempty"`
	// HomeProject names the project at the top of the workspace whose
	// directory its runs get as HOME, by its folder; "" unsets it. On an
	// update, left out keeps it.
	HomeProject *string `json:"home_project,omitempty"`
}
