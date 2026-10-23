package api

import (
	"time"
)

// ArtifactEntry is a file or a directory of the artifact tree, as a
// listing of a directory shows it: a file's size and the run that wrote
// its latest version, or the user who wrote it with no run.
type ArtifactEntry struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Dir      bool      `json:"dir"`
	ByteSize int64     `json:"byte_size"`
	Modified time.Time `json:"modified"`
	RunID    string    `json:"run_id,omitempty"`
	Folder   string    `json:"folder,omitempty"`
	Job      string    `json:"job,omitempty"`
	UserName string    `json:"user_name,omitempty"`
}
