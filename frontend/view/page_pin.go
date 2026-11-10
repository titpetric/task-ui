package view

import (
	"strings"

	"github.com/titpetric/task-ui/model"
)

// BrowseID returns the path of the directory a browse page shows, as a pin names it: the folder's ID and the directory.
func (p *Page) BrowseID() string {
	return p.Folder.ID + "/" + strings.TrimSuffix(p.Browse, "/")
}

// PinOf returns the user's pin of a folder, or of a job of it with no args, nil for none.
func (p *Page) PinOf(folderID, job, dir string) *model.Pin {
	for i, pin := range p.Pins {
		if pin.Folder == folderID && pin.Job == job && pin.Dir == dir && (pin.Args == "" || pin.Args == "[]") {
			return &p.Pins[i]
		}
	}
	return nil
}

// PinItem returns the [...] item that pins what names, a folder or a job of it, or unpins it when pinned.
func (p *Page) PinItem(folderID, job, dir, what string) MenuItem {
	key := folderID + "|" + dir + "|" + job
	attrs := []string{"data-pin-folder", folderID, "data-pin-job", job, "data-pin-dir", dir, "data-pin-key", key}
	label, title := "Pin "+what, pinTitle(job, false)
	if pin := p.PinOf(folderID, job, dir); pin != nil {
		label, title = "Unpin "+what, pinTitle(job, true)
		attrs = append(attrs, "data-pin-id", pin.ID)
	}
	item, _ := NewMenuItem("pin", label, append(attrs, "title", title)...)
	return item
}

// pinTitle says what a pin of a folder or of a job does, and what
// removing it does: a folder brings its project to the sidebar, which
// lists workspaces and projects, and a job is marked out on the page of
// the folder that runs it.
func pinTitle(job string, pinned bool) string {
	switch {
	case job != "" && pinned:
		return "Stop marking it out on the folder's page"
	case job != "":
		return "Mark it out on the folder's page"
	case pinned:
		return "Take its project out of the sidebar, unless the sidebar lists it anyway"
	}
	return "Show its project in the sidebar"
}
