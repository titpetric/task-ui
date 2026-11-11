package model

import (
	"strings"
)

// Default is the folder of the default workspace: the root itself, which
// holds every project that is in no named workspace, its built-in jobs
// and older runs of the Claude jobs, internal/claude-note, claude-prompt
// and claude. No checkout, project or other workspace takes the name.
const Default = "internal"

// Legacy is the name the default workspace had before it was internal.
// Stored rows moved to Default; a URL or an argument that names it is
// read as Default.
const Legacy = "_"

// Canonical returns the folder ID with its first segment _ read as the default workspace.
func Canonical(folder string) string {
	top, rest, found := strings.Cut(folder, "/")
	if top != Legacy {
		return folder
	}
	if !found {
		return Default
	}
	return Default + "/" + rest
}

// IsReserved reports whether a folder ID starts with a name task-ui keeps: the default workspace's internal, or the _ it had.
func IsReserved(folder string) bool {
	top, _, _ := strings.Cut(folder, "/")
	return top == Default || top == Legacy
}
