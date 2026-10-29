package folder

import (
	"path"
	"strings"
)

// repository is the directory task-ui lists but never goes into: the git
// repository of a checkout.
const repository = ".git"

// Sealed reports whether task-ui lists an entry but keeps it closed: .git, or a dot file that is not a dir.
func Sealed(name string, dir bool) bool {
	return name == repository || (!dir && strings.HasPrefix(name, ".") && name != "." && name != "..")
}

// SealedPath reports whether a slash path names a sealed entry or has .git among its segments.
// It takes .git/config, sub/.env and .claude/.secret, and leaves .claude/CLAUDE.md open; dir marks the last entry a directory.
func SealedPath(rel string, dir bool) bool {
	segments := strings.Split(strings.TrimPrefix(path.Clean("/"+rel), "/"), "/")
	for i, segment := range segments {
		if Sealed(segment, dir || i < len(segments)-1) {
			return true
		}
	}
	return false
}
