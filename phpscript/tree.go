package phpscript

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/artifactstore"
)

// ReadLimit is the most a module reads out of one artifact. A report is
// prose and a schema is a few kilobytes; anything past this is a file to
// link rather than a file to render, and the page links it either way.
const ReadLimit = 256 << 10

// Tree is the instance's artifact tree as a module reads it: the files
// its runs left, and the text of one.
//
// A module cannot reach the tree on its own. The filesystem boundary
// roots every path a script opens at the module's own directory, which
// is the point: a plugin reads its own files and the artifacts through
// a binding that knows what the artifacts are.
//
// Reading only. A module shows what a run wrote; the runs write.
type Tree struct {
	Store *artifactstore.Store
}

// NewTree returns the reader over the instance's artifact tree. A nil
// store is an instance with none, and both calls answer empty.
func NewTree(store *artifactstore.Store) *Tree {
	return &Tree{Store: store}
}

// Files is every file of the tree, by path, as a module lists them.
// Directories are walked and not listed: a module shows files.
func (t *Tree) Files() []string {
	if t == nil || t.Store == nil {
		return nil
	}
	root, err := t.Store.Path("")
	if err != nil {
		return nil
	}
	var files []string
	err = filepath.WalkDir(root, func(at string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			// The tree is a git repository and its history is not an
			// artifact anybody asked for.
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, at)
		if err != nil {
			return nil
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil
	}
	sort.Strings(files)
	return files
}

// Read is the text of one artifact, and an empty string for a path that
// is not there, is a directory, or holds more than ReadLimit.
//
// The path is read from the top of the tree. A path leaving it is
// refused the way a module's own paths are.
func (t *Tree) Read(rel string) string {
	if t == nil || t.Store == nil {
		return ""
	}
	root, err := t.Store.Path("")
	if err != nil {
		return ""
	}
	clean, err := Inside(root, rel)
	if err != nil {
		return ""
	}
	at := filepath.Join(root, clean)
	info, err := os.Stat(at)
	if err != nil || info.IsDir() || info.Size() > ReadLimit {
		return ""
	}
	data, err := os.ReadFile(at)
	if err != nil {
		return ""
	}
	return string(data)
}

// IsImage reports whether a path is one the page embeds rather than
// links, by its extension: a schema diagram is read on the page and a
// SQL file is a link.
func IsImage(rel string) bool {
	switch strings.ToLower(path.Ext(rel)) {
	case ".svg", ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return true
	}
	return false
}
