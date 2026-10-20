package agent

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// ask runs claude -p on the prompt in dir and writes its reply to out.
func (a *Agent) ask(ctx context.Context, dir, prompt string, out io.Writer) error {
	if a.getenv("CLAUDE_CODE_OAUTH_TOKEN") == "" {
		return errNoToken
	}
	path, err := a.claudePath()
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, path, "-p", "--dangerously-skip-permissions", "--output-format", "text", prompt)
	cmd.Dir = dir
	cmd.Env = a.environ()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = a.Stdin, out, a.Stderr
	return cmd.Run()
}
