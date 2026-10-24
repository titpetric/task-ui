package view

import (
	"path"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/folder"
)

// Sidebar is the folder tree of the sidebar: the folders it lists, in
// the catalogue's order, and the pinned jobs each lists besides its own.
// With no pins it lists the folders seen and the directories right
// inside the workspaces' src, less the projects set not to show there,
// the workspaces' homes and the folders inside those. A pin adds
// to that in its place: a pinned folder with its workspace and project
// above it, a pinned job under its folder; a folder listed for its
// pinned jobs alone lists only those, and a pin in a home stays out.
type Sidebar struct {
	Folders []*folder.Folder
	// Pins are the pinned jobs a folder's row does not list, by folder ID.
	Pins map[string][]SidePin
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

// NewSidebar returns the tree of the default folders with the pins placed among them.
func NewSidebar(defaults []*folder.Folder, pins []SidePin, workspaces map[string]int, homes map[string]bool) Sidebar {
	s := Sidebar{Folders: append([]*folder.Folder(nil), defaults...), Pins: map[string][]SidePin{}}
	listed := map[string]*folder.Folder{}
	for _, f := range defaults {
		listed[f.ID] = f
	}
	add := func(f *folder.Folder) *folder.Folder {
		if had, ok := listed[f.ID]; ok {
			return had
		}
		listed[f.ID] = f
		s.Folders = append(s.Folders, f)
		return f
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
		if f := add(&bare); !pin.plain(f) {
			s.Pins[f.ID] = append(s.Pins[f.ID], pin)
		}
	}
	for _, pin := range pins {
		for _, id := range above(pin.Folder.ID, workspaces) {
			add(&folder.Folder{ID: id, Title: path.Base(id)})
		}
	}
	sort.SliceStable(s.Folders, func(i, j int) bool {
		return strings.ToLower(s.Folders[i].ID) < strings.ToLower(s.Folders[j].ID)
	})
	return s
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
