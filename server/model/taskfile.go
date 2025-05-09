package model

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type Taskfile struct {
	Tasks []*Task `json:"tasks"`
}

func LoadTaskfile(filename string) (*Taskfile, error) {
	var result Taskfile

	// Run the `task -l --json` command for the taskfile
	cmd := exec.Command("task", "-t", filename, "-l", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run task command: %w", err)
	}

	// Decode the JSON output into the Taskfile{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON output: %w", err)
	}

	return &result, nil
}

type Task struct {
	Name        string   `json:"name"`
	Description string   `json:"desc"`
	Summary     string   `json:"summary"`
	Aliases     []string `json:"aliases"`

	Internal    bool `json:"internal"`
	Interactive bool `json:"interactive"`

	// These are not being read from task -l;
	// Task --show-summary doesn't support -j; yet?
	Cmds []string `json:"-"`
	Vars []string `json:"-"`
}

type Command struct {
	Name string
	Args []string
}
