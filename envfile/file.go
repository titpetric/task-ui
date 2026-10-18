package envfile

// File is a dotenv file a folder reads, as task-ui manages it.
type File struct {
	Folder string `json:"folder"`
	// Path is the file path relative to the folder, as the source names it.
	Path string `json:"path"`
	// Managed is set once the content was saved from the UI.
	Managed bool `json:"managed"`
	// Exists is set when the file is present in the folder.
	Exists bool `json:"exists"`
	// Content is the managed content, or the folder's file when not managed.
	Content string `json:"-"`
	// Keys are the variable names the content declares.
	Keys []string `json:"keys"`
}

// State names what the folder page shows for the file.
func (f File) State() string {
	switch {
	case f.Managed:
		return "managed"
	case f.Exists:
		return "in folder"
	}
	return "missing"
}
