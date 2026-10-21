package detector

import (
	"fmt"
	"strings"
)

// fileSpec is the subset of an atkins pipeline or a Taskfile that task-ui
// reads. Both share `tasks:` with `desc`, `summary`, `aliases`, `internal`
// and `interactive`; atkins adds `jobs:`, `name:`, `help:`, `include:` and
// `env.include:`, Taskfiles add `dotenv:`. The decoder skips unknown keys,
// so either file parses, and nothing task-ui needs is an extension.
type fileSpec struct {
	Name    string     `yaml:"name"`
	Help    string     `yaml:"help"`
	Include stringList `yaml:"include"`
	Env     envSpec    `yaml:"env"`
	Dotenv  stringList `yaml:"dotenv"`
	// Includes is the Taskfile map of namespace to included Taskfile.
	Includes map[string]taskfileInclude `yaml:"includes"`
	Jobs     map[string]jobSpec         `yaml:"jobs"`
	Tasks    map[string]jobSpec         `yaml:"tasks"`
}

// envSpec is an atkins env declaration: `{vars, include}`.
type envSpec struct {
	Include stringList `yaml:"include"`
}

// taskfileInclude is one Taskfile include: a path, or a mapping with the
// path under taskfile.
type taskfileInclude struct {
	Taskfile string `yaml:"taskfile"`
}

func (i *taskfileInclude) UnmarshalYAML(node *yamlNode) error {
	if node.Kind == yamlScalar {
		i.Taskfile = strings.TrimSpace(node.Value)
		return nil
	}
	type raw taskfileInclude
	return node.Decode((*raw)(i))
}

// stringList accepts one string or a list of strings.
type stringList []string

func (l *stringList) UnmarshalYAML(node *yamlNode) error {
	switch node.Kind {
	case yamlScalar:
		if strings.TrimSpace(node.Value) != "" {
			*l = []string{strings.TrimSpace(node.Value)}
		}
		return nil
	case yamlSequence:
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
	Args        bool       `yaml:"args"`
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
func (j *jobSpec) UnmarshalYAML(node *yamlNode) error {
	switch node.Kind {
	case yamlScalar:
		cmd := strings.TrimSpace(node.Value)
		j.Desc = cmd
		j.Steps = []stepSpec{{Run: cmd, bare: true}}
		return nil
	case yamlSequence:
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

// firstLabel is the label of the job's first labelled step, what atkins
// shows for a job without a desc.
func (j *jobSpec) firstLabel() string {
	for _, step := range j.steps() {
		if label := step.Label(); label != "" {
			return label
		}
	}
	return ""
}

// exec returns the argv of the job's first exec: step, the command the
// arguments of a job that takes them follow; nil without one.
func (j *jobSpec) exec() []string {
	for _, step := range j.steps() {
		if len(step.Exec) > 0 {
			return step.Exec
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
	Exec     []string `yaml:"exec"`
	Task     string   `yaml:"task"`
	Defer    string   `yaml:"defer"`
	Deferred bool     `yaml:"deferred"`
	bare     bool
}

func (s *stepSpec) UnmarshalYAML(node *yamlNode) error {
	if node.Kind == yamlScalar {
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
	case len(s.Exec) > 0:
		label = "exec: " + strings.Join(s.Exec, " ")
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
