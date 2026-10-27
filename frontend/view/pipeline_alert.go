package view

import (
	"strings"

	"github.com/titpetric/task-ui/model"
)

// alertRows is how many rows a pipeline alert shows before it folds the
// rest under Show more.
const alertRows = 5

// PipelineAlert is the alert a page raises for the broken atkins
// pipelines in a place: how many pipelines atkins refuses, the rows it
// shows open and the ones it folds, and the workspaces they are in. A
// place is a folder and every folder inside it, the folders of the
// default workspace for internal, and every folder the user sees for "".
type PipelineAlert struct {
	Pipelines  int
	Shown      []PipelineError
	More       []PipelineError
	Workspaces []string
}

// Rows counts the findings the alert lists.
func (a *PipelineAlert) Rows() int {
	return len(a.Shown) + len(a.More)
}

// PipelineAlert gathers the lint findings of the folders in the place
// id names, nil where there are none.
func (p *Page) PipelineAlert(id string) *PipelineAlert {
	alert := &PipelineAlert{}
	pipelines := map[string]bool{}
	workspaces := map[string]bool{}
	for _, f := range p.Folders {
		if len(f.Errors) == 0 || !p.inPlace(id, f.ID) {
			continue
		}
		workspace := p.workspaceOf(f.ID)
		if !workspaces[workspace] {
			workspaces[workspace] = true
			alert.Workspaces = append(alert.Workspaces, workspace)
		}
		for _, finding := range f.Errors {
			if key := f.ID + "\x00" + finding.Path; !pipelines[key] {
				pipelines[key] = true
				alert.Pipelines++
			}
			row := PipelineError{Folder: f, Finding: finding}
			if len(alert.Shown) < alertRows {
				alert.Shown = append(alert.Shown, row)
			} else {
				alert.More = append(alert.More, row)
			}
		}
	}
	if alert.Pipelines == 0 {
		return nil
	}
	return alert
}

// inPlace reports whether the folder of id is the place or inside it.
func (p *Page) inPlace(place, id string) bool {
	switch place {
	case "":
		return true
	case model.Default:
		return p.workspaceOf(id) == model.Default
	}
	return id == place || strings.HasPrefix(id, place+"/")
}

// workspaceOf returns the workspace a folder is in: the named one its ID
// starts with, or internal for a project at the root.
func (p *Page) workspaceOf(id string) string {
	top, _, _ := strings.Cut(id, "/")
	if _, named := p.Workspaces[top]; named {
		return top
	}
	return model.Default
}
