package view

import (
	"strings"

	"github.com/titpetric/task-ui/folder"
)

// Group is a project in the sidebar: its top folder and the folders
// inside it, in their order. Top is nil for folders whose top is not
// listed.
type Group struct {
	Top    *folder.Folder
	Nested []*folder.Folder
}

// Groups puts the folders, in their order, under their projects.
func Groups(folders []*folder.Folder) []Group {
	var groups []Group
	byTop := map[string]int{}
	for _, f := range folders {
		if f.Depth() == 0 {
			byTop[f.ID] = len(groups)
			groups = append(groups, Group{Top: f})
		}
	}
	for _, f := range folders {
		if f.Depth() == 0 {
			continue
		}
		top, _, _ := strings.Cut(f.ID, "/")
		if i, ok := byTop[top]; ok {
			groups[i].Nested = append(groups[i].Nested, f)
			continue
		}
		byTop[top] = len(groups)
		groups = append(groups, Group{Nested: []*folder.Folder{f}})
	}
	return groups
}
