package folder

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/git"
)

// Entry is one directory under the root, as the settings list it: with or
// without something to run, with or without a git remote.
type Entry struct {
	ID  string
	Dir string
	// Sources are the files the directory's conventions were read from.
	Sources []string
	Jobs    int
	// Checkout is what git knows about the directory, nil when it is not
	// a checkout.
	Checkout *git.Info
	// Error is set when a source does not parse.
	Error string
}

// Entries lists the directories under root.
func Entries(root string, g *git.Git) []Entry {
	entries, _ := os.ReadDir(root)
	result := []Entry{}
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		dir := filepath.Join(root, entry.Name())
		item := Entry{ID: SafeID(entry.Name()), Dir: dir, Checkout: g.Info(dir)}
		if detector.Any(dir) {
			f := Load(item.ID, dir)
			for _, source := range f.Sources {
				item.Sources = append(item.Sources, source.File)
			}
			item.Jobs = len(f.Jobs)
			item.Error = f.Error
		}
		result = append(result, item)
	}
	return result
}
