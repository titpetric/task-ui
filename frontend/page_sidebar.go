package frontend

import (
	"context"
	"os"
	"path"
	"strings"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
)

// sidebar fills the page's folder tree: the folders seen and the workspaces' directories, less hidden and the homes, each workspace's projects cut to its sidebar limit, then the user's pins.
func (h *Handlers) sidebar(ctx context.Context, p *view.Page) {
	hidden, homes := map[string]bool{}, map[string]bool{}
	if h.opts.Projects != nil {
		projects, _ := h.opts.Projects.List(ctx)
		for _, project := range projects {
			if project.SidebarHidden {
				hidden[project.Folder] = true
			}
		}
	}
	hidden[layout.Home], homes[layout.Home] = true, true
	for ws := range p.Workspaces {
		home := layout.HomeID(ws)
		hidden[home], homes[home] = true, true
	}
	side := view.SideDefaults(p.Folders, h.workspaceDirs(p.Workspaces), hidden, p.User.CanSee)
	dirOf := h.dirOf(ctx, side)
	side = h.sideLimits(ctx).Apply(side, p.Workspaces, func(id string) time.Time { return h.lastCommit(dirOf(id)) })
	p.Pinning = h.opts.Pins != nil && p.User != nil
	var pins []view.SidePin
	if p.Pinning {
		p.Pins, _ = h.opts.Pins.List(ctx, p.User.Name)
		for _, pin := range p.Pins {
			if side, ok := sidePin(p, pin); ok {
				pins = append(pins, side)
			}
		}
	}
	tree := view.NewSidebar(side, pins, p.Workspaces, homes)
	p.SideFolders, p.SidePins = tree.Folders, tree.Pins
	p.SideDirty = h.sideDirty(tree.Folders, h.dirOf(ctx, tree.Folders))
	p.FillSide(h.sideRows(ctx))
}

// sideRows is the sidebar's top block as the table holds it.
//
// An instance whose table cannot be read falls back to the fixed list the
// block held before the table existed, so a database that refuses one
// query does not leave the page with no way out of itself.
func (h *Handlers) sideRows(ctx context.Context) []model.Sidebar {
	if h.opts.Sidebar == nil {
		return model.SidebarDefaults()
	}
	stored, err := h.opts.Sidebar.List(ctx)
	if err != nil || len(stored) == 0 {
		return model.SidebarDefaults()
	}
	return stored
}

// lastCommitOf answers the time of a project's last commit, for the
// sidebar's limits: in the directory the catalogue knows for a folder,
// else where the layout puts the project.
func (h *Handlers) lastCommitOf(ctx context.Context, folders []*folder.Folder) func(id string) time.Time {
	dirs := map[string]string{}
	for _, f := range folders {
		if f.Dir != "" {
			dirs[f.ID] = f.Dir
		}
	}
	var named []string
	return func(id string) time.Time {
		dir, ok := dirs[id]
		if !ok {
			if named == nil {
				named = h.workspaceFolders(ctx)
			}
			dir = layout.Dir(h.opts.Root, named, id)
		}
		return h.lastCommit(dir)
	}
}

// lastCommit is the time of the last commit in a directory, through the
// cache both front ends share; zero on an instance with no git.
func (h *Handlers) lastCommit(dir string) time.Time {
	if h.opts.Git == nil {
		return time.Time{}
	}
	return h.opts.Git.Commit(dir)
}

// sideLimits returns how many projects the sidebar lists for each
// workspace that limits them, the root among them.
func (h *Handlers) sideLimits(ctx context.Context) view.SideLimits {
	limits := view.SideLimits{}
	if h.opts.Workspaces == nil {
		return limits
	}
	workspaces, _ := h.opts.Workspaces.List(ctx)
	for _, ws := range workspaces {
		if ws.SidebarLimit > 0 {
			limits[ws.Folder] = int(ws.SidebarLimit)
		}
	}
	return limits
}

// workspaceDirs returns the directories in the src of each named workspace by ID, none without a root.
func (h *Handlers) workspaceDirs(workspaces map[string]int) []string {
	var ids []string
	for ws := range workspaces {
		if h.opts.Root == "" {
			return nil
		}
		entries, _ := os.ReadDir(layout.SrcDir(h.opts.Root, ws))
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				ids = append(ids, ws+"/"+entry.Name())
			}
		}
	}
	return ids
}

// sidePin resolves a pin among the folders the page's user sees: a job
// of a folder, a folder, or a directory inside one; false for one that
// is not there or not seen.
func sidePin(p *view.Page, pin model.Pin) (view.SidePin, bool) {
	f, rest := folder.Resolve(seen(p.Folders), pin.Folder)
	switch {
	case f == nil:
		return view.SidePin{}, false
	case pin.Job != "":
		sub := f.Sub(pin.Dir)
		if rest != "" || sub == nil || sub.Job(pin.Job) == nil {
			return view.SidePin{}, false
		}
		return view.SidePin{Pin: pin, Folder: f, Job: sub.Job(pin.Job)}, true
	case rest != "":
		dir, ok := f.Path(rest)
		if info, err := os.Stat(dir); !ok || err != nil || !info.IsDir() {
			return view.SidePin{}, false
		}
		f = &folder.Folder{ID: pin.Folder, Title: path.Base(pin.Folder)}
	}
	return view.SidePin{Pin: pin, Folder: f}, true
}

// seen is the folders a user sees as a catalogue, for folder.Resolve.
type seen []*folder.Folder

func (s seen) List() []*folder.Folder { return s }

func (s seen) Find(id string) *folder.Folder {
	for _, f := range s {
		if f.ID == id {
			return f
		}
	}
	return nil
}
