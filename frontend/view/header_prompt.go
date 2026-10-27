package view

import (
	"strings"

	"github.com/titpetric/task-ui/model"
)

// HeaderPrompt is where the header's prompt asks on a page: the folder
// and the directory of a folder page, a directory, a job, an env file
// or a project's or workspace's settings; the run of a run's page, in
// its folder and directory; the root, internal, anywhere else. The header
// carries it in data- attributes.
type HeaderPrompt struct {
	// Folder and Dir are the place /api/prompts asks in, internal for the root.
	Folder string
	Dir    string
	// Run is the run a run's page asks about, in Folder and Dir.
	Run string
	// Label names the place in the panel's first line: malta/dockerfiles,
	// run and the start of its ID, or internal.
	Label string
	// Href is the page of the thread: the place's page or the run's.
	Href string
}

// HeaderPrompt returns where the header's prompt asks on the page.
func (p *Page) HeaderPrompt() HeaderPrompt {
	root := HeaderPrompt{Folder: model.Default, Label: model.Default, Href: "/f/" + model.Default}
	switch {
	case p.Run != nil:
		run := p.Run.Run
		if run.IsClaude() {
			return p.headerPlace(run.Folder, run.Dir, root)
		}
		place := p.headerPlace(run.Folder, run.Dir, root)
		short := run.ID
		if len(short) > 4 {
			short = short[:4] + "…"
		}
		place.Run, place.Label, place.Href = run.ID, "run "+short, "/runs/"+run.ID
		return place
	case p.Workspace != nil:
		return p.headerPlace(p.Workspace.Folder, "", root)
	case p.Folder != nil:
		dir := strings.Trim(p.Browse, "/")
		if p.Job != nil && dir == "" {
			dir = p.Job.Dir
		}
		return p.headerPlace(p.Folder.ID, dir, root)
	}
	return root
}

// headerPlace is the header's prompt in a folder and a directory inside
// it; the root for _ or the internal folder of older runs.
func (p *Page) headerPlace(folderID, dir string, root HeaderPrompt) HeaderPrompt {
	if folderID == model.Default && dir == "" || folderID == "" {
		return root
	}
	where := model.Run{Folder: folderID, Dir: dir}.Where()
	href := p.PlaceHref(where)
	if href == "" {
		href = "/f/" + model.Default
	}
	return HeaderPrompt{Folder: folderID, Dir: dir, Label: where, Href: href}
}
