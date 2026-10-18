package folder

import (
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// Config overrides or declares a folder, from its project row.
type Config struct {
	ID          string   `yaml:"id"`
	Path        string   `yaml:"path"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Allow       []string `yaml:"allow"`
	// Command replaces the command of the folder's only source, such as
	// `gmake` for a Makefile. Commands does the same per kind when the
	// folder follows several conventions.
	Command  string            `yaml:"command"`
	Commands map[string]string `yaml:"commands"`
	// Sources limits the conventions read in the folder, by kind:
	// atkins, task, make, npm, composer, script. Empty reads them all.
	Sources []string `yaml:"sources"`
	// Jobs narrows who may run single jobs: job ID to users or roles.
	Jobs map[string][]string `yaml:"jobs"`
	// Hours says when the folder's jobs may run; JobHours does so per job.
	Hours    string            `yaml:"hours"`
	JobHours map[string]string `yaml:"job_hours"`
	// RootOnly lists the folder's own jobs alone: discovery does not go
	// into it. Off, the default, the folders inside it are listed too.
	RootOnly bool `yaml:"root_only"`
}

// dir returns the directory the configuration names, under root.
func (c Config) dir(root string) string {
	dir := c.Path
	if dir == "" {
		dir = filepath.FromSlash(c.ID)
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}
	return dir
}

// Apply puts the configuration on a loaded folder.
func (c Config) Apply(f *Folder) {
	if c.Title != "" {
		f.Title = c.Title
	}
	if c.Description != "" {
		f.Description = c.Description
	}
	if len(c.Allow) > 0 {
		f.Allow = c.Allow
	}
	commands := map[string][]string{}
	for kind, command := range c.Commands {
		commands[kind] = strings.Fields(command)
	}
	if c.Command != "" && len(f.Sources) == 1 {
		commands[f.Sources[0].Kind] = strings.Fields(c.Command)
	}
	for _, source := range f.Sources {
		if command, ok := commands[source.Kind]; ok {
			source.Command = command
			for _, job := range f.SourceJobs(source) {
				job.Command = append(append([]string{}, command...), job.Argv()...)
			}
		}
	}
	for name, allow := range c.Jobs {
		if job := f.Job(name); job != nil {
			job.Allow = allow
		}
	}
	if c.Hours != "" {
		f.Hours = c.Hours
	}
	for name, hours := range c.JobHours {
		if job := f.Job(name); job != nil {
			job.Hours = hours
		}
	}
}

// ConfigOf reads a project row as the configuration of its folder.
func ConfigOf(p *model.Project) Config {
	if p == nil {
		return Config{}
	}
	return Config{
		ID:          p.Folder,
		Title:       p.Title,
		Description: p.Description,
		Allow:       p.AllowList(),
		Commands:    p.CommandMap(),
		Sources:     p.SourceList(),
		Jobs:        p.JobAllowMap(),
		Hours:       p.Hours,
		JobHours:    p.JobHoursMap(),
		RootOnly:    p.RootOnly,
	}
}

// ConfigsOf reads project rows as folder configurations.
func ConfigsOf(projects []model.Project) []Config {
	configs := make([]Config, 0, len(projects))
	for i := range projects {
		configs = append(configs, ConfigOf(&projects[i]))
	}
	return configs
}
