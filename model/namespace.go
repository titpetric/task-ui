package model

import (
	"strings"
)

// Internal is the namespace of the folder task-ui builds for its own host
// commands, the notes agent among them. No checkout or project takes the
// name, and only admins see the folder.
const Internal = "internal"

// Default is the folder of the default workspace: the root itself, which
// holds every project that is in no named workspace. No checkout,
// project or other workspace takes the name.
const Default = "_"

// IsInternal reports whether a folder ID is the internal folder or one
// under it.
func IsInternal(folder string) bool {
	return folder == Internal || strings.HasPrefix(folder, Internal+"/")
}

// IsReserved reports whether a folder ID starts with a name task-ui keeps
// for itself: internal or the default workspace's _.
func IsReserved(folder string) bool {
	top, _, _ := strings.Cut(folder, "/")
	return top == Internal || top == Default
}
