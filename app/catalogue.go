package app

import (
	"context"
	"log/slog"

	"github.com/titpetric/task-ui/agent"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/skills"
	"github.com/titpetric/task-ui/storage"
)

// Catalogue discovers the folders under the root on every call, with the
// project rows applied, so a clone or a configuration change shows at
// once. The internal folder of the host commands comes first.
type Catalogue struct {
	Root     string
	Discover bool
	Projects *storage.ProjectStorage
	// Agent is the notes agent's user, whose job in the internal folder
	// allows it.
	Agent string
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
	return d
}

// List returns the internal folder, then the folders of the root ordered
// by ID. A project row under the internal namespace lists nothing.
func (c *Catalogue) List() []*folder.Folder {
	list := []*folder.Folder{agent.Folder(c.Root, c.Agent)}
	for _, f := range c.discovery().List() {
		if !model.IsInternal(f.ID) {
			list = append(list, f)
		}
	}
	return list
}

// Find returns a folder by ID, or nil.
func (c *Catalogue) Find(id string) *folder.Folder {
	if id == agent.Namespace {
		return agent.Folder(c.Root, c.Agent)
	}
	if model.IsInternal(id) {
		return nil
	}
	return c.discovery().Find(id)
}
