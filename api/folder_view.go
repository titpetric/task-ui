package api

// FolderView is a folder as the jobs listing shows it: what it is, and
// its jobs for the user asking.
type FolderView struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Jobs        []JobView `json:"jobs"`
	// Error says why the folder's files could not be read.
	Error string `json:"error,omitempty"`
}
