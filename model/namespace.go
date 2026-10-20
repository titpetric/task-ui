package model

import (
	"strings"
)

// Internal is the namespace of the folder task-ui builds for its own host
// commands, the notes agent among them. No checkout or project takes the
// name, and only admins see the folder.
const Internal = "internal"

// IsInternal reports whether a folder ID is the internal folder or one
// under it.
func IsInternal(folder string) bool {
	return folder == Internal || strings.HasPrefix(folder, Internal+"/")
}
