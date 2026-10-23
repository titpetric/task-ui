package detector

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// yamlNode is the yaml node the spec types decode from. This file names
// it so the one import of the yaml package in this package is this file.
type yamlNode = yaml.Node

// The node kinds the spec types tell apart.
const (
	yamlScalar   = yaml.ScalarNode
	yamlSequence = yaml.SequenceNode
)

// taskfileNames are the file names go-task discovers.
var taskfileNames = []string{"Taskfile.yml", "taskfile.yml", "Taskfile.yaml", "taskfile.yaml", "Taskfile.dist.yml", "taskfile.dist.yml", "Taskfile.dist.yaml", "taskfile.dist.yaml"}

// pipelineDetector reads a Taskfile: the first of its file names
// present in the folder.
//
// An atkins pipeline is read by asking atkins, in atkins.go. go-task has
// no equivalent listing this reads, so its file is parsed here.
type pipelineDetector struct {
	kind  string
	names []string
}

func (d pipelineDetector) Kind() string { return d.kind }

func (d pipelineDetector) Detect(dir string) *Source {
	for _, name := range d.names {
		if isFile(dir, name) {
			return d.detectFile(dir, name)
		}
	}
	return nil
}

// isFile reports whether a folder holds a regular file by that name.
func isFile(dir, name string) bool {
	info, err := os.Stat(filepath.Join(dir, name))
	return err == nil && !info.IsDir()
}

// detectFile reads one pipeline file: a Taskfile lists the jobs with a
// desc, as `task -l` does, and lists no internal or hidden job.
func (d pipelineDetector) detectFile(dir, name string) *Source {
	source := &Source{Kind: d.kind, File: name, Command: defaultCommand(d.kind, name)}

	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		source.Error = err.Error()
		return source
	}
	var spec fileSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		source.Error = fmt.Sprintf("%s: %s", name, err)
		return source
	}

	source.Name = spec.Name
	source.Description = strings.TrimSpace(spec.Help)
	source.Includes = spec.Include
	for _, namespace := range sortedNames(spec.Includes) {
		if path := spec.Includes[namespace].Taskfile; path != "" {
			source.Includes = appendUnique(source.Includes, path)
		}
	}
	source.EnvFiles = appendUnique(nil, spec.Env.Include...)
	source.EnvFiles = appendUnique(source.EnvFiles, spec.Dotenv...)

	names := make([]string, 0, len(spec.Tasks))
	for name := range spec.Tasks {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		job := spec.Tasks[name]
		desc := job.Desc
		if job.Internal || desc == "" || (job.Show != nil && !*job.Show) {
			continue
		}
		j := &Target{
			Name:        name,
			Description: desc,
			Summary:     strings.TrimSpace(job.Summary),
			Aliases:     job.Aliases,
			Interactive: job.Interactive,
			DependsOn:   job.DependsOn,
			EnvFiles:    job.Env.Include,
		}
		for _, dep := range job.Deps {
			if dep.Task != "" {
				j.DependsOn = append(j.DependsOn, dep.Task)
			} else if dep.bare {
				j.DependsOn = append(j.DependsOn, dep.Run)
			}
		}
		for _, step := range job.steps() {
			if label := step.Label(); label != "" {
				j.Steps = append(j.Steps, label)
			}
		}
		source.EnvFiles = appendUnique(source.EnvFiles, j.EnvFiles...)
		source.Targets = append(source.Targets, j)
	}
	return source
}

func sortedNames(m map[string]taskfileInclude) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func appendUnique(list []string, items ...string) []string {
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		found := false
		for _, have := range list {
			if have == item {
				found = true
				break
			}
		}
		if !found {
			list = append(list, item)
		}
	}
	return list
}

// defaultCommand returns the command for a pipeline file kind.
func defaultCommand(kind, file string) []string {
	if kind == "atkins" {
		return []string{"atkins", "-f", file}
	}
	return []string{"task", "-t", file}
}
