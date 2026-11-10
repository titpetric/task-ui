package view

import (
	"path"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// SidePin is a pin of the user as the sidebar places it: the pin and the
// folder it names, with the job for a pin of one.
type SidePin struct {
	Pin    model.Pin
	Folder *folder.Folder
	// Job is the pinned job, nil for a folder pin.
	Job *folder.Job
}

// SideDefaults returns what the sidebar lists with no pins, the folders under hidden left out.
func SideDefaults(folders []*folder.Folder, dirs []string, hidden map[string]bool, sees func(string) bool) []*folder.Folder {
	listed := map[string]bool{}
	var side []*folder.Folder
	for _, f := range folders {
		if !under(hidden, f.ID) {
			side = append(side, f)
			listed[f.ID] = true
		}
	}
	for _, id := range dirs {
		if !listed[id] && !under(hidden, id) && sees(id) {
			side = append(side, &folder.Folder{ID: id, Title: path.Base(id)})
		}
	}
	return side
}

// NewSidebar returns the folders the sidebar lists, in the catalogue's
// order, with the pins placed among them.
//
// With no pins it lists the folders seen and the directories right inside
// the workspaces' src, less the projects set not to show there and
// everything at or under a workspace's home. A pin adds to that in its
// place: the folder it names, with the workspace and the project above it,
// so a pin of a folder no default listed draws that folder's project. The
// sidebar draws workspaces and projects, and a folder inside one is a row
// of the project's page. A pin in a home stays out.
func NewSidebar(defaults []*folder.Folder, pins []SidePin, workspaces map[string]int, homes map[string]bool) []*folder.Folder {
	folders := append([]*folder.Folder(nil), defaults...)
	listed := map[string]*folder.Folder{}
	for _, f := range defaults {
		listed[f.ID] = f
	}
	add := func(f *folder.Folder) {
		if _, ok := listed[f.ID]; ok {
			return
		}
		listed[f.ID] = f
		folders = append(folders, f)
	}
	kept := pins[:0:0]
	for _, pin := range pins {
		if !under(homes, pin.Folder.ID) {
			kept = append(kept, pin)
		}
	}
	pins = kept
	for _, pin := range pins {
		if pin.Job == nil {
			add(pin.Folder)
		}
	}
	for _, pin := range pins {
		if pin.Job == nil {
			continue
		}
		bare := *pin.Folder
		bare.Jobs = nil
		add(&bare)
	}
	for _, pin := range pins {
		for _, id := range above(pin.Folder.ID, workspaces) {
			add(&folder.Folder{ID: id, Title: path.Base(id)})
		}
	}
	sort.SliceStable(folders, func(i, j int) bool {
		return strings.ToLower(folders[i].ID) < strings.ToLower(folders[j].ID)
	})
	return folders
}

// above returns the workspace and the project a folder is in, those above it.
func above(id string, workspaces map[string]int) []string {
	parts := strings.Split(id, "/")
	levels := 1
	if _, ok := workspaces[parts[0]]; ok {
		levels = 2
	}
	var ids []string
	for i := 1; i <= levels && i < len(parts); i++ {
		ids = append(ids, strings.Join(parts[:i], "/"))
	}
	return ids
}

// under reports whether set holds a folder or one above it.
func under(set map[string]bool, id string) bool {
	for {
		if set[id] {
			return true
		}
		i := strings.LastIndex(id, "/")
		if i < 0 {
			return false
		}
		id = id[:i]
	}
}
