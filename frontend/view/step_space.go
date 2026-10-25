package view

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/folder"
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

// StepSpaces lists _, while it exists or a step folder lies at the root, then by rank each named workspace holding a folder of StepFolders.
func (p *Page) StepSpaces() []StepSpace {
	titles := map[string]string{}
	root := false
	for _, ws := range p.WorkspaceList {
		titles[ws.Folder] = ws.Title
		root = root || ws.Folder == model.Default
	}
	seen := map[string]bool{}
	for _, f := range p.StepFolders {
		if ws := p.workspaceOf(f.ID); ws != model.Default {
			seen[ws] = true
		} else {
			root = true
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
	spaces := []StepSpace{}
	if root {
		spaces = append(spaces, StepSpace{ID: model.Default, Title: "Default"})
		if title := titles[model.Default]; title != "" {
			spaces[0].Title = title
		}
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

// StepRel returns a folder ID relative to its workspace, the key a step
// keeps when it moves between workspaces: the ID itself in the default
// workspace, the rest after the workspace's folder in a named one, and
// "." for the root of either.
func (p *Page) StepRel(id string) string {
	ws := p.workspaceOf(id)
	if ws == model.Default {
		if id == model.Default || id == "" {
			return "."
		}
		return id
	}
	if rel := strings.TrimPrefix(strings.TrimPrefix(id, ws), "/"); rel != "" {
		return rel
	}
	return "."
}

// StepLabel returns what the folder picker shows of a folder ID: StepRel, the root for ".".
func (p *Page) StepLabel(id string) string {
	if rel := p.StepRel(id); rel != "." {
		return rel
	}
	return "the root"
}

// StepFolderIn returns the folder ID of a relative folder in a workspace.
func StepFolderIn(ws, rel string) string {
	switch {
	case rel == "." && ws == model.Default:
		return model.Default
	case rel == ".":
		return ws
	case ws == model.Default:
		return rel
	}
	return ws + "/" + rel
}

// stepFolder returns the step folder of an ID, nil when the user has none such with a job.
func (p *Page) stepFolder(id string) *folder.Folder {
	for _, f := range p.StepFolders {
		if f.ID == id && len(f.Jobs) > 0 {
			return f
		}
	}
	return nil
}

// stepHasJob reports whether a folder holds a job a command step picks; the root's are its host jobs.
func (p *Page) stepHasJob(id, job string) bool {
	if id == model.Default {
		for _, j := range p.HostJobs {
			if j.ID == job {
				return true
			}
		}
		return false
	}
	if f := p.stepFolder(id); f != nil {
		for _, j := range f.Jobs {
			if j.ID == job {
				return true
			}
		}
	}
	return false
}

// spaceTitle returns the title the workspace picker shows for a workspace.
func (p *Page) spaceTitle(ws string) string {
	for _, s := range p.StepSpaces() {
		if s.ID == ws {
			return s.Title
		}
	}
	return ws
}

// StepMove moves the steps to a workspace by their relative folders, clearing and counting what it lacks.
func (p *Page) StepMove(steps []Step, ws string) (moved []Step, missed int) {
	title := p.spaceTitle(ws)
	moved = make([]Step, 0, len(steps))
	for _, s := range steps {
		s.Missing = ""
		kind := s.StepKind()
		if kind == model.StepWorkflow || s.Folder == "" {
			moved = append(moved, s)
			continue
		}
		rel := p.StepRel(s.Folder)
		to := StepFolderIn(ws, rel)
		name := p.StepLabel(s.Folder)
		switch {
		case kind != model.StepJob:
			if rel != "." && p.stepFolder(to) == nil {
				s.Missing = fmt.Sprintf("%s is not in %s.", name, title)
				to = StepFolderIn(ws, ".")
			}
			s.Folder = to
		case to != model.Default && p.stepFolder(to) == nil:
			s.Missing = fmt.Sprintf("%s is not in %s.", name, title)
			s.Folder, s.Job, s.Args = "", "", ""
		case !p.stepHasJob(to, s.Job):
			s.Missing = fmt.Sprintf("%s has no job %s in %s.", name, s.Job, title)
			s.Folder, s.Job, s.Args = to, "", ""
		default:
			s.Folder = to
		}
		if s.Missing != "" {
			missed++
		}
		moved = append(moved, s)
	}
	return moved, missed
}

// StepSpaceOf returns the workspace of a folder ID, _ for the root and the projects at its top.
func (p *Page) StepSpaceOf(id string) string {
	return p.workspaceOf(id)
}

// StepMoveItems returns the edited workflow's menu items that move its steps to each other workspace.
func (p *Page) StepMoveItems() []MenuItem {
	spaces := p.StepSpaces()
	if p.Workflow == nil || len(spaces) < 2 {
		return nil
	}
	items := []MenuItem{{Sep: true}}
	for _, ws := range spaces {
		item, _ := NewMenuItem("folder", "Move to "+ws.Title, "href", "/workflows/"+p.Workflow.ID+"/edit?workspace="+url.QueryEscape(ws.ID))
		items = append(items, item)
	}
	return append(items, MenuItem{Sep: true})
}

// StepSpaceHas reports whether a workspace is one the steps may move to.
func (p *Page) StepSpaceHas(ws string) bool {
	for _, s := range p.StepSpaces() {
		if s.ID == ws {
			return true
		}
	}
	return false
}
