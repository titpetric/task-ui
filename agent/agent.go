// Package agent holds task-ui's host commands, the commands of the Claude
// built-ins every folder has: Claude Code on a note from the Notes board,
// @claude-note, on a prompt, @claude-prompt, and in a terminal, @claude.
// The server runs them as `task-ui agent note|prompt|shell`.
package agent

import (
	"context"
	"fmt"
	"io"
)

// Agent runs one host command with an environment and a terminal: the
// run's own, when task-ui starts the job.
type Agent struct {
	// Env is the environment the command reads and Claude Code runs
	// with, os.Environ() for a run.
	Env    []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Run runs the command the arguments name. A Claude Code that exits
// non-zero returns its *exec.ExitError, so the run ends with its code.
func (a *Agent) Run(ctx context.Context, args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "note":
			return a.Note(ctx)
		case "prompt":
			return a.Prompt(ctx)
		case "shell":
			return a.Shell(args[1:])
		}
	}
	return fmt.Errorf("usage: task-ui agent note|prompt|shell [args]")
}
