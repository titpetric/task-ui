package folder

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/skills"
)

// maxDepth bounds how far discovery descends below the root.
const maxDepth = 6

// skipped are the directories discovery never looks into.
var skipped = map[string]bool{"node_modules": true, "vendor": true, "testdata": true}

// Discovery finds the folders under a root: the root itself when it holds
// something to run, and every directory below it that does, as deep as
// they go. Under a folder with a pipeline file Discovery lists only folders with
// a pipeline file of their own: a Go tree's Makefiles are the pipeline's
// business. Configured folders override the discovered ones by ID, and
// Discovery adds a configured folder whose path the walk did not find.
type Discovery struct {
	Root     string
	Discover bool
	Config   []Config
	// Probe is every folder's Probe, for the skills atkins applies in it.
	Probe *skills.Prober
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
		sections(root, c)
	}

	result := make([]*Folder, 0, len(c.byID))
	for _, f := range c.byID {
		f.Probe = d.Probe
		result = append(result, f)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].ID) < strings.ToLower(result[j].ID)
	})
	return result
}

// configured loads the folder a configuration names: a directory, a
// single file, or a directory with nothing to run of its own, which holds
// folders that do. A path that is missing carries an error, so the
// configuration shows as wrong.
func configured(root string, cfg Config) *Folder {
	dir := cfg.dir(root)
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
	case isDir(dir):
		f = &Folder{ID: id, Title: id, Dir: dir}
	default:
		f = &Folder{ID: id, Title: id, Dir: dir, Error: "no such directory: " + dir}
	}
	cfg.Apply(f)
	return f
}

// isDir reports whether a path is a directory.
func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// isFile reports whether a path is a file.
func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// discover adds the root when it has something to run, and walks the
// tree below it, one level into a shallow project. A directory of the root
// named internal or _ lists nothing: the names are task-ui's own. The walk
// goes into every configured folder, even one a checkout ignores, such
// as a project cloned inside a workspace.
func (d *Discovery) discover(root string, c *catalogue) {
	if !c.seen(root) && detector.Any(root) {
		if f := Load(SafeID(filepath.Base(root)), root); runnable(f) {
			c.add(f)
		}
	}
	shallow, keep := map[string]bool{}, map[string]bool{}
	for _, cfg := range d.Config {
		keep[cfg.dir(root)] = true
		if cfg.Shallow {
			shallow[cfg.dir(root)] = true
		}
	}
	w := walker{shallow: shallow, keep: keep, found: func(id, dir string) {
		if c.seen(dir) || model.IsReserved(id) {
			return
		}
		if f := Load(id, dir); runnable(f) {
			c.add(f)
		}
	}}
	w.walk(root, "", 1, detector.HasPipeline(root), gitIgnored(root, nil), false)
}

// sections adds the top level folder of every listed folder below one
// that has nothing to run itself, such as a checkout of many apps: the
// sidebar lists the folders inside it under it.
func sections(root string, c *catalogue) {
	ids := make([]string, 0, len(c.byID))
	for id := range c.byID {
		ids = append(ids, id)
	}
	for _, id := range ids {
		top, _, nested := strings.Cut(id, "/")
		if !nested || model.IsReserved(top) {
			continue
		}
		if dir := filepath.Join(root, top); !c.seen(dir) {
			c.add(Load(top, dir))
		}
	}
}

// runnable reports whether a discovered folder has a job to run, or a
// source that failed to read, which the folder page shows.
func runnable(f *Folder) bool {
	for _, source := range f.Sources {
		if len(source.Targets) > 0 || source.Error != "" {
			return true
		}
	}
	return false
}

// gitIgnored adds the directories a git checkout at dir ignores to
// ignored: build copies, caches, dependencies. dir that is not a checkout
// adds nothing, and nor does a failing git.
func gitIgnored(dir string, ignored map[string]bool) map[string]bool {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		return ignored
	}
	out, err := exec.Command("git", "-C", dir, "ls-files", "-z", "--others", "--ignored", "--exclude-standard", "--directory").Output()
	if err != nil {
		return ignored
	}
	merged := map[string]bool{}
	for path := range ignored {
		merged[path] = true
	}
	for _, path := range bytes.Split(out, []byte{0}) {
		if p := strings.TrimSuffix(string(path), "/"); p != "" {
			merged[filepath.Join(dir, p)] = true
		}
	}
	return merged
}

// walker visits the directories below the root and reports the ones
// with something to run to found. Directories a git checkout ignores are
// not visited unless keep holds them, and the walk goes one level into
// a directory in shallow, a shallow project, and no deeper.
type walker struct {
	shallow, keep map[string]bool
	found         func(id, dir string)
}

// walk visits the directories below dir, in name order. Below a pipeline
// folder walk reports only a folder with a pipeline file, but goes on
// through the rest. last stops it under the directories of dir: dir is a
// shallow project.
func (w walker) walk(dir, id string, depth int, underPipeline bool, ignored map[string]bool, last bool) {
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
		if ignored[sub] && !w.keep[sub] {
			continue
		}
		ignored := gitIgnored(sub, ignored)
		subID := SafeID(name)
		if id != "" {
			subID = id + "/" + subID
		}
		pipeline := detector.HasPipeline(sub)
		if pipeline || (!underPipeline && detector.Any(sub)) {
			w.found(subID, sub)
		}
		if last {
			continue
		}
		w.walk(sub, subID, depth+1, underPipeline || pipeline, ignored, w.shallow[sub])
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
