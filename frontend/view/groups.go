package view

import (
	"strings"

	"github.com/titpetric/task-ui/folder"
)

// Group is a project in the sidebar: its top folder and the folders
// inside it, in their order. Top is nil for folders whose top is not
// listed; ID names the top either way. A workspace is a group whose
// projects stand under it in Projects, each a group of its own, and
// whose Nested is empty.
type Group struct {
	ID        string
	Top       *folder.Folder
	Nested    []*folder.Folder
	Workspace bool
	Projects  []Group
}

// Groups puts the folders, in their order, under their projects, and
// the projects of each workspace, by its folder, under the workspace.
func Groups(folders []*folder.Folder, workspaces map[string]bool) []Group {
	return group(folders, 0, workspaces)
}

// group groups the folders by their first depth+1 segments: a folder
// that deep is a group's top, a deeper one sits in its group.
func group(folders []*folder.Folder, depth int, workspaces map[string]bool) []Group {
	var groups []Group
	byTop := map[string]int{}
	for _, f := range folders {
		if f.Depth() == depth {
			byTop[f.ID] = len(groups)
			groups = append(groups, Group{ID: f.ID, Top: f, Workspace: workspaces[f.ID]})
		}
	}
	for _, f := range folders {
		if f.Depth() == depth {
			continue
		}
		top := strings.Join(strings.SplitN(f.ID, "/", depth+2)[:depth+1], "/")
		i, ok := byTop[top]
		if !ok {
			i = len(groups)
			byTop[top] = i
			groups = append(groups, Group{ID: top, Workspace: workspaces[top]})
		}
		groups[i].Nested = append(groups[i].Nested, f)
	}
	for i := range groups {
		if groups[i].Workspace {
			groups[i].Projects = group(groups[i].Nested, depth+1, nil)
			groups[i].Nested = nil
		}
	}
	return groups
}
