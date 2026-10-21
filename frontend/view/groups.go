package view

import (
	"sort"
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
// The projects at the root come first, in the folders' order; the
// workspaces follow, by their rank in workspaces and then by folder.
func Groups(folders []*folder.Folder, workspaces map[string]int) []Group {
	groups := group(folders, 0, workspaces)
	sort.SliceStable(groups, func(i, j int) bool {
		a, b := groups[i], groups[j]
		if a.Workspace != b.Workspace {
			return b.Workspace
		}
		if !a.Workspace {
			return false
		}
		if ra, rb := workspaces[a.ID], workspaces[b.ID]; ra != rb {
			return ra < rb
		}
		return a.ID < b.ID
	})
	return groups
}

// group groups the folders by their first depth+1 segments: a folder
// that deep is a group's top, a deeper one sits in its group. A group
// stands where its first folder does, its top or, with the top not
// listed, the first folder inside it.
func group(folders []*folder.Folder, depth int, workspaces map[string]int) []Group {
	var groups []Group
	byTop := map[string]int{}
	for _, f := range folders {
		top := f.ID
		if f.Depth() != depth {
			top = strings.Join(strings.SplitN(f.ID, "/", depth+2)[:depth+1], "/")
		}
		i, ok := byTop[top]
		if !ok {
			i = len(groups)
			byTop[top] = i
			_, workspace := workspaces[top]
			groups = append(groups, Group{ID: top, Workspace: workspace})
		}
		if f.Depth() == depth {
			groups[i].Top = f
			continue
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
