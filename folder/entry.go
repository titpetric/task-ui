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
	ID  string `json:"id"`
	Dir string `json:"dir"`
	// Sources are the files that hold the directory's conventions.
	Sources []string `json:"sources"`
	Jobs    int      `json:"jobs"`
	// Checkout is what git knows about the directory, nil for a directory
	// that is not a checkout.
	Checkout *git.Info `json:"checkout,omitempty"`
	// Error holds the parse error when a source does not parse.
	Error string `json:"error,omitempty"`
}

// Entries lists the directories under root. Jobs counts the jobs of the
// directory and of the folders discovered inside it.
func Entries(root string, g *git.Git) []Entry {
	entries, _ := os.ReadDir(root)
	nested := map[string]int{}
	for _, f := range (&Discovery{Root: root, Discover: true}).List() {
		if top, _, ok := strings.Cut(f.ID, "/"); ok {
			nested[top] += len(f.Jobs)
		}
	}
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
		item.Jobs += nested[item.ID]
		result = append(result, item)
	}
	return result
}
