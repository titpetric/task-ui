package api

// FolderView is a folder as the jobs listing shows it: its title and
// description, and its jobs for the user asking.
type FolderView struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Jobs        []JobView `json:"jobs"`
	// Error says why reading the folder's files failed.
	Error string `json:"error,omitempty"`
}
