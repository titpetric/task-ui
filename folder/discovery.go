package folder

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/detector"
)

// maxDepth bounds how far discovery descends below the root.
const maxDepth = 6

// skipped are the directories discovery never looks into.
var skipped = map[string]bool{"node_modules": true, "vendor": true, "testdata": true}

// Discovery finds the folders under a root: the root itself when it holds
// something to run, and every directory below it that does, as deep as
// they go. Under a folder with a pipeline file only folders with a
// pipeline file of their own are listed: a Go tree's Makefiles are the
// pipeline's business. Configured folders are applied over the
// discovered ones by ID, and added when their path was not discovered.
type Discovery struct {
	Root     string
	Discover bool
	Config   []Config
}

// List returns the folders in tree order: a folder, then the folders
// inside it.
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
			dir = filepath.FromSlash(cfg.ID)
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
		walk(root, "", 1, detector.HasPipeline(root), func(id, dir string) {
			if _, seen := byDir[dir]; !seen {
				add(Load(id, dir))
			}
		})
	}

	result := make([]*Folder, 0, len(byID))
	for _, f := range byID {
		result = append(result, f)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].ID) < strings.ToLower(result[j].ID)
	})
	return result
}

// walk visits the directories below dir, in name order, and reports the
// ones with something to run. Below a pipeline folder only a folder with
// a pipeline file is reported, but the walk goes on through the rest.
func walk(dir, id string, depth int, underPipeline bool, found func(id, dir string)) {
	if depth > maxDepth {
		return
	}
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || strings.HasPrefix(name, ".") || skipped[name] {
			continue
		}
		sub := filepath.Join(dir, name)
		subID := SafeID(name)
		if id != "" {
			subID = id + "/" + subID
		}
		pipeline := detector.HasPipeline(sub)
		if pipeline || (!underPipeline && detector.Any(sub)) {
			found(subID, sub)
		}
		walk(sub, subID, depth+1, underPipeline || pipeline, found)
	}
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
