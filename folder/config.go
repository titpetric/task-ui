package folder

import (
	"strings"
)

// Config overrides or declares a folder in task-ui.yml.
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
}
