package api

// ArtifactFileView is a file of the artifact tree as the artifact window
// shows it: one version, as an image, markdown rendered to HTML, text, or
// none of these, and the versions the runs wrote, newest first.
type ArtifactFileView struct {
	Path        string            `json:"path"`
	Version     string            `json:"version"`
	Kind        string            `json:"kind"`
	ContentType string            `json:"content_type"`
	ByteSize    int64             `json:"byte_size"`
	HTML        string            `json:"html,omitempty"`
	Text        string            `json:"text,omitempty"`
	Versions    []ArtifactVersion `json:"versions"`
}
