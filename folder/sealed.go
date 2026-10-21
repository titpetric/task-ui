package folder

import (
	"path"
	"strings"
)

// repository is the directory task-ui lists but never goes into: the git
// repository of a checkout.
const repository = ".git"

// Sealed reports whether an entry of a checkout is one task-ui lists but
// does not open, show, run or go into: .git, and any dot file, a file
// whose name starts with a dot. dir says the entry is a directory, or a
// symlink to one: a dot directory other than .git opens, the dot files
// in it stay sealed.
func Sealed(name string, dir bool) bool {
	return name == repository || (!dir && strings.HasPrefix(name, ".") && name != "." && name != "..")
}

// SealedPath reports whether a slash path names a sealed entry or goes
// through .git: it takes .git/config, sub/.env and .claude/.secret, and
// leaves .claude/CLAUDE.md. dir says the last entry is a directory; the
// ones before it are.
func SealedPath(rel string, dir bool) bool {
	segments := strings.Split(strings.TrimPrefix(path.Clean("/"+rel), "/"), "/")
	for i, segment := range segments {
		if Sealed(segment, dir || i < len(segments)-1) {
			return true
		}
	}
	return false
}
