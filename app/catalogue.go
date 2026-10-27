package app

import (
	"context"
	"log/slog"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/skills"
	"github.com/titpetric/task-ui/storage"
)

// Catalogue discovers the folders under the root on every call, with the
// project rows applied, so a clone or a configuration change shows at
// once. Discovery lists the root under no ID, so Find answers _ with the
// src of the default workspace: the folder of the root's built-in jobs,
// @shell, @claude and @host:<host>.
type Catalogue struct {
	Root     string
	Discover bool
	Projects *storage.ProjectStorage
	// Workspaces are the workspace rows, whose directories discovery
	// treats as roots.
	Workspaces *storage.WorkspaceStorage
	// Skills asks atkins for the skills of a folder's directory, for
	// every folder the catalogue lists.
	Skills *skills.Prober

	mu      sync.Mutex
	every   []*folder.Folder
	everyAt time.Time
}

// everyTTL is how long the workflow editor's listing of every folder
// with a job lasts: the walk goes into every checkout.
const everyTTL = 30 * time.Second

// NewCatalogue returns a catalogue over the root and the project rows.
func NewCatalogue(root string, discover bool, projects *storage.ProjectStorage) *Catalogue {
	return &Catalogue{Root: root, Discover: discover, Projects: projects}
}

// discovery builds the discovery of the moment.
func (c *Catalogue) discovery() *folder.Discovery {
	return c.discover(false)
}

// discover builds the discovery of the moment, listing every folder with a job for every.
func (c *Catalogue) discover(every bool) *folder.Discovery {
	d := &folder.Discovery{Root: c.Root, Discover: c.Discover, Probe: c.Skills, Every: every}
	if c.Projects != nil {
		projects, err := c.Projects.List(context.Background())
		if err != nil {
			slog.Error("listing projects", "err", err)
		}
		d.Config = folder.ConfigsOf(projects)
	}
	if c.Workspaces != nil {
		workspaces, err := c.Workspaces.List(context.Background())
		if err != nil {
			slog.Error("listing workspaces", "err", err)
		}
		d.Workspaces = folder.WorkspaceConfigsOf(workspaces)
	}
	return d
}

// List returns the folders of the root ordered by ID.
func (c *Catalogue) List() []*folder.Folder {
	return c.discovery().List()
}

// Find returns a folder by ID, or nil; internal that no folder takes is
// the root, in its src while the default workspace exists, else the root
// itself. The _ it was named before reads as internal.
func (c *Catalogue) Find(id string) *folder.Folder {
	id = model.Canonical(id)
	d := c.discovery()
	if f := d.Find(id); f != nil || id != model.Default {
		return f
	}
	root := &folder.Folder{ID: model.Default, Title: model.Default, Dir: c.Root, Hours: d.HoursOf(model.Default)}
	if slices.ContainsFunc(d.Workspaces, func(cfg folder.Config) bool { return cfg.ID == model.Default }) {
		root.Dir = layout.SrcDir(c.Root, model.Default)
	}
	return root
}

// Home returns the home of the workspace a directory under the root is
// in, the HOME atkins lists its skills with there; "" for a directory
// outside the root or with no workspaces known.
func (c *Catalogue) Home(dir string) string {
	if c.Workspaces == nil {
		return ""
	}
	root, err := filepath.Abs(c.Root)
	if err != nil {
		return ""
	}
	named, _ := c.Workspaces.Folders(context.Background())
	workspace, ok := layout.Workspace(root, named, dir)
	if !ok {
		return ""
	}
	return layout.HomeDir(root, workspace)
}

// Every returns each folder with a job, whatever the settings hide, cached for everyTTL.
func (c *Catalogue) Every() []*folder.Folder {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.every == nil || time.Since(c.everyAt) > everyTTL {
		c.every, c.everyAt = c.discover(true).List(), time.Now()
	}
	return c.every
}

// Steps returns the catalogue of workflow steps, which lists and finds by Every too.
func (c *Catalogue) Steps() folder.Catalogue {
	return steps{c}
}

// steps is the catalogue of a workflow's steps: the workflow editor
// picks from Every, and the workflow runner finds a folder the
// workspace and project settings hide.
type steps struct{ c *Catalogue }

// List returns every folder with a job.
func (s steps) List() []*folder.Folder {
	return s.c.Every()
}

// Find returns a folder of the catalogue, else one of Every, or nil.
func (s steps) Find(id string) *folder.Folder {
	if f := s.c.Find(id); f != nil {
		return f
	}
	for _, f := range s.c.Every() {
		if f.ID == id {
			return f
		}
	}
	return nil
}
