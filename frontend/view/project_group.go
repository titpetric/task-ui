package view

import (
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// ProjectGroup is a block of the Projects page: a workspace and the
// projects in it, the default workspace's the folders at the top of
// the root, less the workspaces themselves, which the Workspaces page
// lists. ProjectGroups makes them from the entries under the root, in
// the workspaces' order, the root first; a workspace without a project
// keeps a group without rows. The projects table draws each group's
// entries.
type ProjectGroup struct {
	// Workspace is the workspace's folder, _ for the default one.
	Workspace string
	Entries   []folder.Entry
}

// Default reports whether the group is the root's.
func (g ProjectGroup) Default() bool {
	return g.Workspace == model.Default
}

// Title is the block's heading: Projects for the root, the workspace's folder otherwise.
func (g ProjectGroup) Title() string {
	if g.Default() {
		return "Projects"
	}
	return g.Workspace
}

// Note is what the heading says after the title: where the projects are.
func (g ProjectGroup) Note(root string) string {
	if g.Default() {
		return "in " + root + "/src"
	}
	return "inside " + g.Workspace + "/src/"
}

// Empty is the line of a group without projects.
func (g ProjectGroup) Empty() string {
	if g.Default() {
		return "No projects yet: add one from a git repository, or put a folder with jobs under src."
	}
	return "No projects in " + g.Workspace + " yet."
}

// ProjectGroups returns a group per workspace, in their order, each with its entries, the top-level ones the root's when _ is among them.
func ProjectGroups(entries []folder.Entry, workspaces []model.Workspace) []ProjectGroup {
	groups := []ProjectGroup{}
	at := map[string]int{}
	for _, ws := range workspaces {
		if _, ok := at[ws.Folder]; ok {
			continue
		}
		at[ws.Folder] = len(groups)
		groups = append(groups, ProjectGroup{Workspace: ws.Folder})
	}
	for _, entry := range entries {
		top, _, nested := strings.Cut(entry.ID, "/")
		if !nested {
			if _, named := at[top]; named && top != model.Default {
				continue
			}
			top = model.Default
		}
		if i, ok := at[top]; ok {
			groups[i].Entries = append(groups[i].Entries, entry)
		}
	}
	return groups
}
