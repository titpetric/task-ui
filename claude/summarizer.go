// Package claude runs Claude Code for task-ui's own uses, such as the
// summary of a command's output on the inventory's hosts.
package claude

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ErrNoToken reports a summary without a Claude Code token to sign in with.
var ErrNoToken = errors.New("the Claude Code integration has no token for you")

// Summarizer runs claude -p over a document and returns its answer. It runs
// with the default permissions, so Claude reads the document and answers;
// it runs no tools on its own.
type Summarizer struct {
	// Command is the Claude Code binary, claude unless set.
	Command string
	// Timeout bounds one summary.
	Timeout time.Duration
}

// Summarize sends the document on stdin with the prompt, signed in with
// the integration variables in env, with system, when not empty, after
// Claude's own system prompt, and returns what Claude writes. The sign-in
// is a Claude Code token, or OpenRouter's auth token and base URL.
func (s *Summarizer) Summarize(ctx context.Context, env map[string]string, system, prompt, document string) (string, error) {
	signedIn := env["CLAUDE_CODE_OAUTH_TOKEN"] != "" || (env["ANTHROPIC_AUTH_TOKEN"] != "" && env["ANTHROPIC_BASE_URL"] != "")
	if !signedIn {
		return "", ErrNoToken
	}
	run, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	command := s.Command
	if command == "" {
		command = "claude"
	}
	args := []string{"-p", "--output-format", "text"}
	if system != "" {
		args = append(args, "--append-system-prompt", system)
	}
	cmd := exec.CommandContext(run, command, append(args, prompt)...)
	cmd.Env = os.Environ()
	for name, value := range env {
		cmd.Env = append(cmd.Env, name+"="+value)
	}
	cmd.Stdin = strings.NewReader(document)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	cmd.WaitDelay = time.Second
	if err := cmd.Run(); err != nil {
		if run.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("claude ran past %s", s.Timeout)
		}
		return "", fmt.Errorf("claude: %w: %s", err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String()), nil
}
