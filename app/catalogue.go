package app

import (
	"context"
	"log/slog"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/skills"
	"github.com/titpetric/task-ui/storage"
)

// Catalogue discovers the folders under the root on every call, with the
// project rows applied, so a clone or a configuration change shows at
// once.
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
}

// NewCatalogue returns a catalogue over the root and the project rows.
func NewCatalogue(root string, discover bool, projects *storage.ProjectStorage) *Catalogue {
	return &Catalogue{Root: root, Discover: discover, Projects: projects}
}

// discovery builds the discovery of the moment.
func (c *Catalogue) discovery() *folder.Discovery {
	d := &folder.Discovery{Root: c.Root, Discover: c.Discover, Probe: c.Skills}
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

// Find returns a folder by ID, or nil.
func (c *Catalogue) Find(id string) *folder.Folder {
	return c.discovery().Find(id)
}
