package api

// FileView is a file of a folder's checkout as the preview window shows
// it: an image, markdown rendered to HTML, text coloured as the source
// window colours it, or none of these. The window shows text up to its
// bound, and Truncated says when the file goes on.
type FileView struct {
	Folder      string `json:"folder"`
	Path        string `json:"path"`
	Kind        string `json:"kind"`
	ContentType string `json:"content_type"`
	ByteSize    int64  `json:"byte_size"`
	HTML        string `json:"html,omitempty"`
	Text        string `json:"text,omitempty"`
	Truncated   bool   `json:"truncated,omitempty"`
}
