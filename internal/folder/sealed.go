package folder

import (
	"path"
	"strings"
)

// repository is the git directory task-ui lists but never goes into.
const repository = ".git"

// configFiles are the dot files task-ui opens: project configuration, and nothing a secret would be written in.
var configFiles = map[string]bool{
	".gitignore":      true,
	".gitattributes":  true,
	".dockerignore":   true,
	".editorconfig":   true,
	".npmignore":      true,
	".prettierignore": true,
	".eslintignore":   true,
	".prettierrc":     true,
	".eslintrc":       true,
	".golangci.yml":   true,
	".mdox.yml":       true,
	".splint.yml":     true,
	".vale.ini":       true,
}

// Sealed reports whether task-ui lists an entry but keeps it closed: .git, or a dot file that is not a config file or a dir.
func Sealed(name string, dir bool) bool {
	if name == repository {
		return true
	}
	if dir || !strings.HasPrefix(name, ".") || name == "." || name == ".." {
		return false
	}
	return !configFiles[name]
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
