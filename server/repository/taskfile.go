package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/titpetric/task-ui/server/model"
)

// AtkinsNames are the file names atkins discovers, in its order.
var AtkinsNames = []string{".atkins.yml", ".atkins.yaml", "atkins.yml", "atkins.yaml"}

// TaskfileNames are the file names go-task discovers.
var TaskfileNames = []string{"Taskfile.yml", "taskfile.yml", "Taskfile.yaml", "taskfile.yaml", "Taskfile.dist.yml", "taskfile.dist.yml", "Taskfile.dist.yaml", "taskfile.dist.yaml"}

// FindTaskfile returns the task file in dir, atkins files first.
func FindTaskfile(dir string) (string, bool) {
	for _, name := range append(append([]string{}, AtkinsNames...), TaskfileNames...) {
		path := filepath.Join(dir, name)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, true
		}
	}
	return "", false
}

// KindOf returns "atkins" for an atkins file name and "task" otherwise.
func KindOf(path string) string {
	base := filepath.Base(path)
	for _, name := range AtkinsNames {
		if base == name {
			return "atkins"
		}
	}
	return "task"
}

// fileSpec is the subset of an atkins pipeline or a Taskfile that task-ui
// reads. Both formats share `tasks:` with `desc`, `summary`, `aliases`,
// `internal` and `interactive`; atkins adds `jobs:`, `name:` and `help:`.
// Unknown keys are ignored, so either file parses.
type fileSpec struct {
	Name  string             `yaml:"name"`
	Help  string             `yaml:"help"`
	UI    uiSpec             `yaml:"ui"`
	Jobs  map[string]jobSpec `yaml:"jobs"`
	Tasks map[string]jobSpec `yaml:"tasks"`
}

// uiSpec is the optional `ui:` section of a task file.
type uiSpec struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Allow       []string `yaml:"allow"`
	Command     string   `yaml:"command"`
}

type jobSpec struct {
	Desc        string   `yaml:"desc"`
	Summary     string   `yaml:"summary"`
	Aliases     []string `yaml:"aliases"`
	Internal    bool     `yaml:"internal"`
	Interactive bool     `yaml:"interactive"`
	Show        *bool    `yaml:"show"`
	Allow       []string `yaml:"allow"`
}

// UnmarshalYAML accepts the short forms: a job given as a single command
// string, or as a list of commands.
func (j *jobSpec) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		j.Desc = strings.TrimSpace(node.Value)
		return nil
	case yaml.SequenceNode:
		var cmds []any
		if err := node.Decode(&cmds); err != nil {
			return err
		}
		if len(cmds) > 0 {
			if s, ok := cmds[0].(string); ok {
				j.Desc = strings.TrimSpace(s)
			}
		}
		return nil
	}
	type raw jobSpec
	return node.Decode((*raw)(j))
}

// LoadProject reads a task file into a project. A file that does not parse
// returns the project with Error set, so it still shows up in the UI.
func LoadProject(id, path string) *model.Project {
	project := &model.Project{
		ID:    id,
		Title: id,
		Dir:   filepath.Dir(path),
		File:  filepath.Base(path),
		Kind:  KindOf(path),
	}
	project.Command = DefaultCommand(project.Kind, project.File)

	data, err := os.ReadFile(path)
	if err != nil {
		project.Error = err.Error()
		return project
	}

	var spec fileSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		project.Error = fmt.Sprintf("%s: %s", project.File, err)
		return project
	}

	if spec.Name != "" {
		project.Title = spec.Name
	}
	project.Description = strings.TrimSpace(spec.Help)
	if spec.UI.Title != "" {
		project.Title = spec.UI.Title
	}
	if spec.UI.Description != "" {
		project.Description = strings.TrimSpace(spec.UI.Description)
	}
	project.Allow = spec.UI.Allow
	if spec.UI.Command != "" {
		project.Command = strings.Fields(spec.UI.Command)
	}

	jobs := spec.Jobs
	if len(jobs) == 0 {
		jobs = spec.Tasks
	}

	names := make([]string, 0, len(jobs))
	for name := range jobs {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		job := jobs[name]
		if job.Internal || job.Desc == "" || (job.Show != nil && !*job.Show) {
			continue
		}
		project.Runbooks = append(project.Runbooks, &model.Runbook{
			Project:     id,
			Name:        name,
			Description: job.Desc,
			Summary:     strings.TrimSpace(job.Summary),
			Aliases:     job.Aliases,
			Interactive: job.Interactive,
			Allow:       job.Allow,
		})
	}
	return project
}

// DefaultCommand returns the command for a task file kind.
func DefaultCommand(kind, file string) []string {
	if kind == "atkins" {
		return []string{"atkins", "-f", file}
	}
	return []string{"task", "-t", file}
}
