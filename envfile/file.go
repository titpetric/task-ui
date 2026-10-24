package envfile

// File is a dotenv file a user filled for a folder, which task-ui writes
// into the folder before each run.
type File struct {
	Folder string `json:"folder"`
	// Path is the file path relative to the folder.
	Path string `json:"path"`
	// Content is what the user filled, never in JSON.
	Content string `json:"-"`
}
