package api

import "testing"

// A file at the write cap fits the request bound even when every byte of it
// escapes: JSON writes a control character as \u00XX, six bytes, and the
// message escapes with it.
func TestFileRequestBoundCoversEscapedContent(t *testing.T) {
	escaped := maxFileWrite*6 + maxCommitMessage*6
	if maxFileRequest < escaped {
		t.Fatalf("maxFileRequest is %d, short of the %d bytes an escaped file and message take", maxFileRequest, escaped)
	}
}
