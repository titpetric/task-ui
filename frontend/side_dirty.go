package frontend

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/layout"
)

// sideDirty counts the changed and untracked files under each folder
// of the sidebar, by the git status of the checkout it is in; a folder
// outside git, or with none, is left out. Nil without git or a root.
func (h *Handlers) sideDirty(folders []*folder.Folder, dirOf func(id string) string) map[string]int {
	if h.opts.Git == nil || h.opts.Root == "" {
		return nil
	}
	dirty := map[string]int{}
	for _, f := range folders {
		dir := dirOf(f.ID)
		status := h.opts.Git.Status(dir, h.opts.Root)
		if status == nil {
			continue
		}
		rel, err := filepath.Rel(filepath.Join(h.opts.Root, filepath.FromSlash(status.Checkout)), dir)
		if err != nil {
			continue
		}
		rel = strings.TrimPrefix(filepath.ToSlash(rel), ".")
		if n := status.Under(rel); n > 0 {
			dirty[f.ID] = n
		}
	}
	return dirty
}

// dirOf answers the directory of a folder of the sidebar: the one the
// catalogue knows for it, else where the layout puts the project.
func (h *Handlers) dirOf(ctx context.Context, folders []*folder.Folder) func(id string) string {
	dirs := map[string]string{}
	for _, f := range folders {
		if f.Dir != "" {
			dirs[f.ID] = f.Dir
		}
	}
	var named []string
	return func(id string) string {
		if dir, ok := dirs[id]; ok {
			return dir
		}
		if named == nil {
			named = h.workspaceFolders(ctx)
		}
		return layout.Dir(h.opts.Root, named, id)
	}
}
