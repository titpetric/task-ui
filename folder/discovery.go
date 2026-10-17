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

// catalogue collects folders by ID and by directory, the first of each
// winning.
type catalogue struct {
	byID  map[string]*Folder
	byDir map[string]*Folder
}

func (c *catalogue) add(f *Folder) {
	if _, ok := c.byID[f.ID]; ok {
		return
	}
	c.byID[f.ID] = f
	c.byDir[f.Dir] = f
}

func (c *catalogue) seen(dir string) bool {
	_, ok := c.byDir[dir]
	return ok
}

// List returns the folders in tree order: a folder, then the folders
// inside it.
func (d *Discovery) List() []*Folder {
	root, err := filepath.Abs(d.Root)
	if err != nil {
		root = d.Root
	}
	c := &catalogue{byID: map[string]*Folder{}, byDir: map[string]*Folder{}}
	// Configured folders first, so their paths win over discovery.
	for _, cfg := range d.Config {
		c.add(configured(root, cfg))
	}
	if d.Discover {
		d.discover(root, c)
	}

	result := make([]*Folder, 0, len(c.byID))
	for _, f := range c.byID {
		result = append(result, f)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].ID) < strings.ToLower(result[j].ID)
	})
	return result
}

// configured loads the folder a configuration names: a directory, a
// single file, or a path with nothing to run, which is listed with its
// error so the configuration can be seen to be wrong.
func configured(root string, cfg Config) *Folder {
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
	switch {
	case isFile(dir):
		f = LoadFile(id, dir)
	case detector.Any(dir):
		f = Load(id, dir, cfg.Sources...)
	default:
		f = &Folder{ID: id, Title: id, Dir: dir, Error: "nothing to run found in " + dir}
	}
	cfg.Apply(f)
	return f
}

// isFile reports whether a path is a file.
func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// discover adds the root when it has something to run, and walks the
// tree below it.
func (d *Discovery) discover(root string, c *catalogue) {
	if !c.seen(root) && detector.Any(root) {
		c.add(Load(SafeID(filepath.Base(root)), root))
	}
	walk(root, "", 1, detector.HasPipeline(root), func(id, dir string) {
		if !c.seen(dir) {
			c.add(Load(id, dir))
		}
	})
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
