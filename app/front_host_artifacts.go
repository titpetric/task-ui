package app

import (
	"context"
	"path"

	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/phpfront"
)

// Tree is one directory of the artifact tree, as the artifacts page lists it.
// A directory's size is the space the files under it take, which the listing
// measures the way the page's Space does; the entry itself holds no size.
func (h *FrontHost) Tree(_ context.Context, dir string) []phpfront.TreeItem {
	if h.Artifacts == nil {
		return nil
	}
	items := h.Artifacts.List(dir)
	out := make([]phpfront.TreeItem, 0, len(items))
	for _, item := range items {
		each := phpfront.TreeItem{
			Name:    item.Name,
			Dir:     item.Dir,
			Size:    item.Size,
			ModTime: item.ModTime.String(),
		}
		if item.Dir {
			if at, err := h.Artifacts.Path(path.Join(dir, item.Name)); err == nil {
				each.Size = diskusage.Measure(at).Files
			}
		}
		out = append(out, each)
	}
	return out
}

// Space is the space the artifact tree takes, from the meter the Go pages
// read, so the tree is walked once however a reader came to it.
func (h *FrontHost) Space(context.Context) phpfront.Space {
	if h.Artifacts == nil || h.Disk == nil {
		return phpfront.Space{}
	}
	usage := h.Disk.Measure(h.Artifacts.Dir)
	return phpfront.Space{
		Measured: true,
		Files:    usage.Files,
		History:  usage.History,
		Ignored:  usage.Ignored,
	}
}
