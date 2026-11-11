package agent

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/integration/claude"
)

// errNoToken refuses a command whose environment has no Claude Code token.
var errNoToken = errors.New("CLAUDE_CODE_OAUTH_TOKEN or ANTHROPIC_AUTH_TOKEN: set up Settings, Integrations, Claude Code or OpenRouter")

// getenv returns the last value set for name in a.Env, as a process reads it.
func (a *Agent) getenv(name string) string {
	value := ""
	for _, kv := range a.Env {
		if k, v, ok := strings.Cut(kv, "="); ok && k == name {
			value = v
		}
	}
	return value
}

// signedIn reports whether Claude Code has a sign-in: an OAuth token, or
// OpenRouter's auth token with its base URL.
func (a *Agent) signedIn() bool {
	if a.getenv("CLAUDE_CODE_OAUTH_TOKEN") != "" {
		return true
	}
	return a.getenv("ANTHROPIC_AUTH_TOKEN") != "" && a.getenv("ANTHROPIC_BASE_URL") != ""
}

// environ returns a.Env with IS_SANDBOX=1, which skips Claude's permission prompts as root.
// A run with a home also reads the CLAUDE.md of every added directory.
func (a *Agent) environ() []string {
	env := append(append([]string{}, a.Env...), "IS_SANDBOX=1")
	if a.getenv("TASK_UI_HOME") != "" {
		env = append(env, "CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD=1")
	}
	return env
}

// homeArgs returns --add-dir TASK_UI_HOME for a session in dir, so Claude reads that home's CLAUDE.md and skills.
// It returns nil when the home is unset or is dir itself.
func (a *Agent) homeArgs(dir string) []string {
	home := a.getenv("TASK_UI_HOME")
	if home == "" || filepath.Clean(home) == filepath.Clean(dir) {
		return nil
	}
	return []string{"--add-dir", home}
}

// context returns the system prompt every Claude run here appends, in the order system joins it: artifacts, MCP tools, tooling, then claude.Context.
func (a *Agent) context() string {
	artifacts := ""
	if dir := a.getenv("TASK_UI_ARTIFACTS"); dir != "" {
		artifacts = "This run's artifact directory is $TASK_UI_ARTIFACTS, " + dir + ": a file you write there, such as a report or an image, " +
			"moves into task-ui's artifact tree at the same path when the run ends, for the Artifacts page."
	}
	tools := ""
	if a.getenv("TASK_UI_MCP_CONFIG") != "" {
		tools = "task-ui's MCP server, task, is connected for this run as " + a.getenv("TASK_UI_USER") + ", with that user's permissions and no more: " +
			"its tools list and run jobs and workflows, read runs, make and change workflows with get_workflow, create_workflow and update_workflow, " +
			"and read and leave notes. Use them for what the question asks of task-ui; the access ends with the run."
	}
	return system(artifacts, tools, toolset(a.getenv("TASK_UI_TOOLING")), claude.Context(a.getenv("TASK_UI_USER"), a.getenv("TASK_UI_ROLE"), a.getenv("TASK_UI_PERMISSIONS")))
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

// ask runs claude -p on prompt in dir and writes the reply to out.
// It appends system, TASK_UI_MCP_CONFIG and the workspace home when each is set.
func (a *Agent) ask(ctx context.Context, dir, prompt, system string, out io.Writer) error {
	if !a.signedIn() {
		return errNoToken
	}
	path, err := a.claudePath()
	if err != nil {
		return err
	}
	args := append([]string{"-p", "--dangerously-skip-permissions", "--output-format", "text"}, a.homeArgs(dir)...)
	if config := a.getenv("TASK_UI_MCP_CONFIG"); config != "" {
		args = append(args, "--mcp-config", config)
	}
	if system != "" {
		args = append(args, "--append-system-prompt", system)
	}
	cmd := exec.CommandContext(ctx, path, append(args, prompt)...)
	cmd.Dir = dir
	cmd.Env = a.environ()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = a.Stdin, out, a.Stderr
	return cmd.Run()
}
