package app

import (
	"context"
	"log/slog"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/storage"
)

// Catalogue discovers the folders under the root on every call, with the
// project rows applied, so a clone or a configuration change shows at
// once.
type Catalogue struct {
	Root     string
	Discover bool
	Projects *storage.ProjectStorage
}

// NewCatalogue returns a catalogue over the root and the project rows.
func NewCatalogue(root string, discover bool, projects *storage.ProjectStorage) *Catalogue {
	return &Catalogue{Root: root, Discover: discover, Projects: projects}
}

// discovery builds the discovery of the moment.
func (c *Catalogue) discovery() *folder.Discovery {
	d := &folder.Discovery{Root: c.Root, Discover: c.Discover}
	if c.Projects != nil {
		projects, err := c.Projects.List(context.Background())
		if err != nil {
			slog.Error("listing projects", "err", err)
		}
		d.Config = folder.ConfigsOf(projects)
	}
	return d
}

// List returns the folders, ordered by title.
func (c *Catalogue) List() []*folder.Folder {
	return c.discovery().List()
}

// Find returns a folder by ID, or nil.
func (c *Catalogue) Find(id string) *folder.Folder {
	return c.discovery().Find(id)
}
