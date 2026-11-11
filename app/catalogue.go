package app

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/skills"
	"github.com/titpetric/task-ui/internal/folder"
	"github.com/titpetric/task-ui/storage"
)

// Catalogue discovers the folders under the root, with the project rows
// applied, so a clone or a configuration change shows at once. Discovery
// lists the root under no ID, so Find answers _ with the src of the
// default workspace: the folder of the root's built-in jobs, @shell,
// @claude and @host:<host>.
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
	seen    []*folder.Folder
	seenAt  time.Time

	walks atomic.Int64
}

// everyTTL is how long the workflow editor's listing of every folder
// with a job lasts: that walk descends into each checkout.
const everyTTL = 30 * time.Second

// listTTL is how long the catalogue the pages see lasts.
//
// The walk goes into every checkout under the root, and one page takes it
// many times over: the chrome lists the folders, each workflow in the
// sidebar resolves the job of every step through Find to say whether the
// reader may start it, and the ported front end reads those jobs one
// Host::job at a time. A second is shorter than a click and longer than
// a request, so a clone still shows at once and a page walks once.
const listTTL = time.Second

// NewCatalogue returns a catalogue over the root and the project rows.
func NewCatalogue(root string, discover bool, projects *storage.ProjectStorage) *Catalogue {
	return &Catalogue{Root: root, Discover: discover, Projects: projects}
}

// Walks counts the walks of the root since the instance started, the number the listing-cost tests assert to catch an O(n) walk.
func (c *Catalogue) Walks() int64 {
	return c.walks.Load()
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

// List returns the folders of the root ordered by ID, cached for listTTL.
// The slice is the caller's own, because a page sorting one in place rewrites what every page reads for the second after.
func (c *Catalogue) List() []*folder.Folder {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.seen == nil || time.Since(c.seenAt) > listTTL {
		c.walks.Add(1)
		c.seen, c.seenAt = c.discover(false).List(), time.Now()
	}
	return slices.Clone(c.seen)
}

// Find returns a folder by ID, or nil; internal that no folder takes is
// the root, in its src while the default workspace exists, else the root
// itself. The legacy ID _ resolves to internal.
func (c *Catalogue) Find(id string) *folder.Folder {
	id = model.Canonical(id)
	for _, f := range c.List() {
		if f.ID == id {
			return f
		}
	}
	if id != model.Default {
		return nil
	}
	// The root under its own name, which discovery lists under none. The
	// walk here is the one a page never takes: nothing asks Find for
	// internal but the pages of the root's own jobs.
	d := c.discovery()
	root := &folder.Folder{ID: model.Default, Title: model.Default, Dir: c.Root, Hours: d.HoursOf(model.Default)}
	if slices.ContainsFunc(d.Workspaces, func(cfg folder.Config) bool { return cfg.ID == model.Default }) {
		root.Dir = folder.SrcDir(c.Root, model.Default)
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
	workspace, ok := folder.Workspace(root, named, dir)
	if !ok {
		return ""
	}
	return folder.HomeDir(root, workspace)
}

// Every returns each folder with a job, hidden ones included, cached for everyTTL.
func (c *Catalogue) Every() []*folder.Folder {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.every == nil || time.Since(c.everyAt) > everyTTL {
		c.walks.Add(1)
		c.every, c.everyAt = c.discover(true).List(), time.Now()
	}
	return c.every
}

// WorkspaceProjects returns the projects of a workspace, the directories of its src in name order:
// each the folder the catalogue finds, hidden ones too, else a folder of its directory with no jobs but the skills.
func (c *Catalogue) WorkspaceProjects(_ context.Context, workspace string) []*folder.Folder {
	entries, err := os.ReadDir(folder.SrcDir(c.Root, workspace))
	if err != nil {
		return nil
	}
	// The catalogue once, not once per project. Steps().Find answers one
	// ID by building a discovery and walking it, so asking it for every
	// directory here walked every checkout under the root as many times
	// as the workspace has projects: that is seconds on a real root, and
	// it is on every page, because the sidebar lists the workflows and a
	// workflow that runs over a workspace reads its projects to say
	// whether the user may start it.
	known := map[string]*folder.Folder{}
	for _, f := range c.Every() {
		known[f.ID] = f
	}
	// The visible catalogue wins over Every, as Steps().Find prefers it:
	// a folder the project rows configure is that folder, and Every only
	// adds the ones a section or a setting hides.
	for _, f := range c.List() {
		known[f.ID] = f
	}
	var projects []*folder.Folder
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		id := folder.Join(workspace, entry.Name())
		f := known[model.Canonical(id)]
		if f == nil {
			f = &folder.Folder{ID: id, Title: entry.Name(), Dir: folder.ProjectDir(c.Root, workspace, entry.Name()), Probe: c.Skills}
		}
		projects = append(projects, f)
	}
	return projects
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
