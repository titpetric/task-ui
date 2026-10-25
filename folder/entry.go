package folder

import (
	"os"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
)

// Entry is one project directory, under the src of a workspace, as the
// settings list it: with or without something to run, with or without a
// git remote. A checkout inside a named workspace has the ID
// `<workspace>/<name>`, and Jobs counts the jobs of the directory and of
// the folders discovered inside it.
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

// Entries lists the directories of _'s src, when _ is among the workspaces, and the checkouts of each named one's src.
func Entries(root string, g *git.Git, workspaces []model.Workspace) []Entry {
	dirs := map[string]string{}
	ids := []string{}
	add := func(id, dir string) {
		dirs[id] = dir
		ids = append(ids, id)
	}
	d := &Discovery{Root: root, Discover: true}
	for _, row := range workspaces {
		ws := row.Folder
		if ws == model.Default {
			for _, name := range subdirs(layout.SrcDir(root, ws)) {
				if model.IsReserved(SafeID(name)) {
					continue
				}
				add(SafeID(name), layout.ProjectDir(root, ws, name))
			}
			continue
		}
		d.Workspaces = append(d.Workspaces, Config{ID: ws})
		for _, name := range subdirs(layout.SrcDir(root, ws)) {
			dir := layout.ProjectDir(root, ws, name)
			if g.IsRepo(dir) {
				id := ws + "/" + SafeID(name)
				add(id, dir)
				d.Config = append(d.Config, Config{ID: id, Path: dir})
			}
		}
	}
	sort.Strings(ids)

	nested := map[string]int{}
	for _, f := range d.List() {
		for id := f.ID; strings.Contains(id, "/"); {
			id = id[:strings.LastIndex(id, "/")]
			nested[id] += len(f.Jobs)
		}
	}
	result := []Entry{}
	for _, id := range ids {
		dir := dirs[id]
		item := Entry{ID: id, Dir: dir, Checkout: g.Info(dir)}
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

// subdirs returns the names of the directories in dir that are not
// hidden, in name order.
func subdirs(dir string) []string {
	entries, _ := os.ReadDir(dir)
	names := []string{}
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			names = append(names, entry.Name())
		}
	}
	return names
}
