package api

import (
	"time"
)

// ArtifactVersion is one version of a file of the artifact tree: the
// commit that holds it, and the run that wrote it, or the user who wrote
// it with no run.
type ArtifactVersion struct {
	Commit    string     `json:"commit"`
	CreatedAt *time.Time `json:"created_at"`
	Folder    string     `json:"folder"`
	Job       string     `json:"job"`
	RunID     string     `json:"run_id"`
	UserName  string     `json:"user_name,omitempty"`
	ByteSize  int64      `json:"byte_size"`
}
