package view

import (
	"sort"

	"github.com/titpetric/task-ui/model"
)

// StepSpace is a workspace the workflow editor's steps pick before a
// folder: a workflow belongs to no workspace, so each step names the
// workspace of the folder it runs in. ID is the workspace's folder, _
// for the default one; Title is what the picker shows.
type StepSpace struct {
	ID    string
	Title string
}

// StepSpaces lists the default workspace, then by rank each one holding a folder of StepFolders.
func (p *Page) StepSpaces() []StepSpace {
	titles := map[string]string{}
	for _, ws := range p.WorkspaceList {
		titles[ws.Folder] = ws.Title
	}
	seen := map[string]bool{}
	for _, f := range p.StepFolders {
		if ws := p.workspaceOf(f.ID); ws != model.Default {
			seen[ws] = true
		}
	}
	named := make([]string, 0, len(seen))
	for ws := range seen {
		named = append(named, ws)
	}
	sort.Slice(named, func(i, j int) bool {
		a, b := named[i], named[j]
		if ra, rb := p.Workspaces[a], p.Workspaces[b]; ra != rb {
			return ra < rb
		}
		return a < b
	})
	spaces := []StepSpace{{ID: model.Default, Title: "Default"}}
	if title := titles[model.Default]; title != "" {
		spaces[0].Title = title
	}
	for _, ws := range named {
		title := titles[ws]
		if title == "" {
			title = ws
		}
		spaces = append(spaces, StepSpace{ID: ws, Title: title})
	}
	return spaces
}

// StepSpaceOf returns the workspace of a folder ID, _ for the root and the projects at its top.
func (p *Page) StepSpaceOf(id string) string {
	return p.workspaceOf(id)
}
