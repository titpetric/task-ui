package folder

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/detector"
)

// Discovery finds the folders under a root: the root itself when it holds
// something to run, and every direct subdirectory that does. Configured
// folders are applied over the discovered ones by ID, and added when their
// path was not discovered.
type Discovery struct {
	Root     string
	Discover bool
	Config   []Config
}

// List returns the folders ordered by title.
func (d *Discovery) List() []*Folder {
	root, err := filepath.Abs(d.Root)
	if err != nil {
		root = d.Root
	}

	byID := map[string]*Folder{}
	byDir := map[string]*Folder{}
	add := func(f *Folder) {
		if _, ok := byID[f.ID]; ok {
			return
		}
		byID[f.ID] = f
		byDir[f.Dir] = f
	}

	// Configured folders first, so their paths win over discovery.
	for _, cfg := range d.Config {
		dir := cfg.Path
		if dir == "" {
			dir = cfg.ID
		}
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(root, dir)
		}
		id := cfg.ID
		if id == "" {
			id = SafeID(filepath.Base(dir))
		}

		var f *Folder
		if info, err := os.Stat(dir); err == nil && !info.IsDir() {
			f = LoadFile(id, dir)
		} else if detector.Any(dir) {
			f = Load(id, dir, cfg.Sources...)
		} else {
			f = &Folder{ID: id, Title: id, Dir: dir, Error: "nothing to run found in " + dir}
		}
		cfg.Apply(f)
		add(f)
	}

	if d.Discover {
		if _, seen := byDir[root]; !seen && detector.Any(root) {
			add(Load(SafeID(filepath.Base(root)), root))
		}
		entries, _ := os.ReadDir(root)
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			dir := filepath.Join(root, entry.Name())
			if _, seen := byDir[dir]; seen {
				continue
			}
			if detector.Any(dir) {
				add(Load(SafeID(entry.Name()), dir))
			}
		}
	}

	result := make([]*Folder, 0, len(byID))
	for _, f := range byID {
		result = append(result, f)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Title) < strings.ToLower(result[j].Title)
	})
	return result
}

// Find returns a folder by ID, or nil.
func (d *Discovery) Find(id string) *Folder {
	for _, f := range d.List() {
		if f.ID == id {
			return f
		}
	}
	return nil
}
