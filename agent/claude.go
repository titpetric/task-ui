package agent

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/claude"
)

// errNoToken stops a command without a Claude Code token to sign in with.
var errNoToken = errors.New("CLAUDE_CODE_OAUTH_TOKEN: set up Settings, Integrations, Claude Code, and name this user in Who")

// getenv returns a variable of the environment, the last one set winning
// as it does for a process.
func (a *Agent) getenv(name string) string {
	value := ""
	for _, kv := range a.Env {
		if k, v, ok := strings.Cut(kv, "="); ok && k == name {
			value = v
		}
	}
	return value
}

// environ is the environment Claude Code runs with. IS_SANDBOX lets it
// skip the permission prompts as root in the task container.
func (a *Agent) environ() []string {
	return append(append([]string{}, a.Env...), "IS_SANDBOX=1")
}

// context is what every Claude run here reads after the instructions
// of its own: where the run's artifacts go, then claude.Context, the
// user of the run with the role and what the user may do, from
// TASK_UI_USER, TASK_UI_ROLE and TASK_UI_PERMISSIONS, and what task-ui is.
func (a *Agent) context() string {
	artifacts := ""
	if dir := a.getenv("TASK_UI_ARTIFACTS"); dir != "" {
		artifacts = "This run's artifact directory is $TASK_UI_ARTIFACTS, " + dir + ": a file you write there, such as a report or an image, " +
			"moves into task-ui's artifact tree at the same path when the run ends, for the Artifacts page."
	}
	return system(artifacts, claude.Context(a.getenv("TASK_UI_USER"), a.getenv("TASK_UI_ROLE"), a.getenv("TASK_UI_PERMISSIONS")))
}

// claudePath finds claude in the PATH of the environment.
func (a *Agent) claudePath() (string, error) {
	for _, dir := range filepath.SplitList(a.getenv("PATH")) {
		if dir == "" {
			continue
		}
		path := filepath.Join(dir, "claude")
		if info, err := os.Stat(path); err == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
			return path, nil
		}
	}
	return "", errors.New("claude: not found in PATH")
}

// ask runs claude -p on the prompt in dir and writes its reply to out,
// with system, when not empty, after Claude's own system prompt.
func (a *Agent) ask(ctx context.Context, dir, prompt, system string, out io.Writer) error {
	if a.getenv("CLAUDE_CODE_OAUTH_TOKEN") == "" {
		return errNoToken
	}
	path, err := a.claudePath()
	if err != nil {
		return err
	}
	args := []string{"-p", "--dangerously-skip-permissions", "--output-format", "text"}
	if system != "" {
		args = append(args, "--append-system-prompt", system)
	}
	cmd := exec.CommandContext(ctx, path, append(args, prompt)...)
	cmd.Dir = dir
	cmd.Env = a.environ()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = a.Stdin, out, a.Stderr
	return cmd.Run()
}
