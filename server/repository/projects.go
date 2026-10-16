package repository

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/server/model"
)

// ProjectConfig overrides or declares a project in task-ui.yml.
type ProjectConfig struct {
	ID          string   `yaml:"id"`
	Path        string   `yaml:"path"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Allow       []string `yaml:"allow"`
	Command     string   `yaml:"command"`
	// Jobs narrows who may run single jobs: job name to users or roles.
	Jobs map[string][]string `yaml:"jobs"`
}

// Projects discovers the projects under a root folder: the root itself when
// it holds a task file, and every direct subfolder that does. Projects from
// the configuration are applied over the discovered ones by ID, and added
// when their path was not discovered.
type Projects struct {
	Root     string
	Discover bool
	Config   []ProjectConfig
}

var unsafeID = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// SafeID turns a folder name into a URL safe project ID.
func SafeID(name string) string {
	id := strings.Trim(unsafeID.ReplaceAllString(name, "-"), "-.")
	if id == "" {
		return "default"
	}
	return id
}

// List returns the projects ordered by title.
func (p *Projects) List() []*model.Project {
	root, err := filepath.Abs(p.Root)
	if err != nil {
		root = p.Root
	}

	byID := map[string]*model.Project{}
	byDir := map[string]*model.Project{}
	add := func(project *model.Project) {
		if _, ok := byID[project.ID]; ok {
			return
		}
		byID[project.ID] = project
		byDir[project.Dir] = project
	}

	// Configured projects first, so their paths win over discovery.
	for _, cfg := range p.Config {
		dir := cfg.Path
		if dir == "" {
			dir = cfg.ID
		}
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(root, dir)
		}
		id := cfg.ID
		if id == "" {
			id = SafeID(filepath.Base(dir))
		}

		var project *model.Project
		if file, ok := FindTaskfile(dir); ok {
			project = LoadProject(id, file)
		} else if info, err := os.Stat(dir); err == nil && !info.IsDir() {
			project = LoadProject(id, dir)
		} else {
			project = &model.Project{ID: id, Title: id, Dir: dir, Error: "no task file found in " + dir}
		}
		apply(project, cfg)
		add(project)
	}

	if p.Discover {
		if file, ok := FindTaskfile(root); ok {
			if _, seen := byDir[root]; !seen {
				add(LoadProject(SafeID(filepath.Base(root)), file))
			}
		}
		entries, _ := os.ReadDir(root)
		for _, entry := range entries {
			if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			dir := filepath.Join(root, entry.Name())
			if _, seen := byDir[dir]; seen {
				continue
			}
			if file, ok := FindTaskfile(dir); ok {
				add(LoadProject(SafeID(entry.Name()), file))
			}
		}
	}

	result := make([]*model.Project, 0, len(byID))
	for _, project := range byID {
		result = append(result, project)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Title) < strings.ToLower(result[j].Title)
	})
	return result
}

// Find returns a project by ID, or nil.
func (p *Projects) Find(id string) *model.Project {
	for _, project := range p.List() {
		if project.ID == id {
			return project
		}
	}
	return nil
}

func apply(project *model.Project, cfg ProjectConfig) {
	if cfg.Title != "" {
		project.Title = cfg.Title
	}
	if cfg.Description != "" {
		project.Description = cfg.Description
	}
	if len(cfg.Allow) > 0 {
		project.Allow = cfg.Allow
	}
	if cfg.Command != "" {
		project.Command = strings.Fields(cfg.Command)
	}
	for name, allow := range cfg.Jobs {
		if job := project.Job(name); job != nil {
			job.Allow = allow
		}
	}
}
