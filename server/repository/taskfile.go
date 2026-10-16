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

// FindTaskfile returns the pipeline file in dir, atkins files first.
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
// reads. Both share `tasks:` with `desc`, `summary`, `aliases`, `internal`
// and `interactive`; atkins adds `jobs:`, `name:`, `help:`, `include:` and
// `env.include:`, Taskfiles add `dotenv:`. Unknown keys are ignored, so
// either file parses, and nothing task-ui needs is an extension.
type fileSpec struct {
	Name    string             `yaml:"name"`
	Help    string             `yaml:"help"`
	Include stringList         `yaml:"include"`
	Env     envSpec            `yaml:"env"`
	Dotenv  stringList         `yaml:"dotenv"`
	Jobs    map[string]jobSpec `yaml:"jobs"`
	Tasks   map[string]jobSpec `yaml:"tasks"`
}

// envSpec is an atkins env declaration: `{vars, include}`.
type envSpec struct {
	Include stringList `yaml:"include"`
}

// stringList accepts one string or a list of strings.
type stringList []string

func (l *stringList) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		if strings.TrimSpace(node.Value) != "" {
			*l = []string{strings.TrimSpace(node.Value)}
		}
		return nil
	case yaml.SequenceNode:
		var items []string
		if err := node.Decode(&items); err != nil {
			return err
		}
		*l = items
		return nil
	}
	return nil
}

type jobSpec struct {
	Desc        string     `yaml:"desc"`
	Summary     string     `yaml:"summary"`
	Aliases     []string   `yaml:"aliases"`
	Internal    bool       `yaml:"internal"`
	Interactive bool       `yaml:"interactive"`
	Show        *bool      `yaml:"show"`
	Run         string     `yaml:"run"`
	Cmd         string     `yaml:"cmd"`
	Steps       []stepSpec `yaml:"steps"`
	Cmds        []stepSpec `yaml:"cmds"`
	DependsOn   stringList `yaml:"depends_on"`
	Deps        []stepSpec `yaml:"deps"`
	Env         envSpec    `yaml:"env"`
}

// UnmarshalYAML accepts the short forms: a job given as a single command
// string, or as a list of commands.
func (j *jobSpec) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		cmd := strings.TrimSpace(node.Value)
		j.Desc = cmd
		j.Steps = []stepSpec{{Run: cmd, bare: true}}
		return nil
	case yaml.SequenceNode:
		var steps []stepSpec
		if err := node.Decode(&steps); err != nil {
			return err
		}
		j.Steps = steps
		if len(steps) > 0 {
			j.Desc = steps[0].Label()
		}
		return nil
	}
	type raw jobSpec
	if err := node.Decode((*raw)(j)); err != nil {
		return err
	}
	if len(j.Steps) == 0 && len(j.Cmds) == 0 {
		if cmd := strings.TrimSpace(firstOf(j.Run, j.Cmd)); cmd != "" {
			j.Steps = []stepSpec{{Run: cmd, bare: true}}
		}
	}
	return nil
}

func (j *jobSpec) steps() []stepSpec {
	if len(j.Steps) > 0 {
		return j.Steps
	}
	return j.Cmds
}

// stepSpec is one step: a bare command, or a mapping naming a command or
// a job, with an optional name and desc for the label.
type stepSpec struct {
	Name     string   `yaml:"name"`
	Desc     string   `yaml:"desc"`
	Run      string   `yaml:"run"`
	Cmd      string   `yaml:"cmd"`
	Cmds     []string `yaml:"cmds"`
	Task     string   `yaml:"task"`
	Defer    string   `yaml:"defer"`
	Deferred bool     `yaml:"deferred"`
	bare     bool
}

func (s *stepSpec) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		s.Run = strings.TrimSpace(node.Value)
		s.bare = true
		return nil
	}
	type raw stepSpec
	if err := node.Decode((*raw)(s)); err != nil {
		return err
	}
	if s.Defer != "" {
		s.Run = s.Defer
		s.Deferred = true
	}
	return nil
}

// Label renders the step the way atkins lists it: the desc, then the
// name, then the command with its kind as a prefix.
func (s *stepSpec) Label() string {
	if s.Desc != "" {
		return s.Desc
	}
	if s.Name != "" {
		return s.Name
	}
	label := ""
	switch {
	case s.Task != "":
		label = "task: " + s.Task
	case s.Run != "":
		if strings.Contains(strings.TrimSpace(s.Run), "\n") {
			label = "run: <script>"
		} else if s.bare {
			label = strings.TrimSpace(s.Run)
		} else {
			label = "run: " + strings.TrimSpace(s.Run)
		}
	case s.Cmd != "":
		label = "cmd: " + strings.TrimSpace(s.Cmd)
	case len(s.Cmds) > 0:
		label = fmt.Sprintf("cmds: <%d commands>", len(s.Cmds))
	}
	if s.Deferred && label != "" {
		label = "defer: " + strings.TrimPrefix(label, "run: ")
	}
	return label
}

func firstOf(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// LoadProject reads a pipeline file into a folder. A file that does not
// parse returns the folder with Error set, so it still shows up in the UI.
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
	project.Includes = spec.Include
	project.EnvFiles = appendUnique(nil, spec.Env.Include...)
	project.EnvFiles = appendUnique(project.EnvFiles, spec.Dotenv...)

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
		j := &model.Job{
			Project:     id,
			Name:        name,
			Description: job.Desc,
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
		project.EnvFiles = appendUnique(project.EnvFiles, j.EnvFiles...)
		project.Jobs = append(project.Jobs, j)
	}
	return project
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

// DefaultCommand returns the command for a pipeline file kind.
func DefaultCommand(kind, file string) []string {
	if kind == "atkins" {
		return []string{"atkins", "-f", file}
	}
	return []string{"task", "-t", file}
}
