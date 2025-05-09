package model

import (
	"fmt"
	"os"
	"sort"

	"github.com/goccy/go-yaml"
)

type Taskfile struct {
	Tasks TaskMap `yaml:"tasks"`
}

type TaskMap map[string]*Task

func (t TaskMap) Keys() []string {
	keys := make([]string, 0, len(t))
	for name := range t {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}

func (t TaskMap) Values() []*Task {
	keys := t.Keys()
	result := make([]*Task, 0, len(keys))
	for _, key := range keys {
		result = append(result, t[key])
	}
	return result
}

func (t TaskMap) Item(name string) *Task {
	v, _ := t[name]
	return v
}

func LoadTaskfile(filename string) (*Taskfile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var result Taskfile
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	for k, v := range result.Tasks {
		v.Name = k
	}

	return &result, nil
}

type Task struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"desc"`
	Summary     string   `yaml:"summary"`
	Aliases     []string `yaml:"aliases"`

	Internal    bool `yaml:"internal"`
	Interactive bool `yaml:"interactive"`
}

type Command struct {
	Name string
	Args []string
}
