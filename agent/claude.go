package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// signedIn reports whether Claude Code has an OAuth token, or OpenRouter's auth token and base URL.
func (a *Agent) signedIn() bool {
	if a.getenv("CLAUDE_CODE_OAUTH_TOKEN") != "" {
		return true
	}
	return a.getenv("ANTHROPIC_AUTH_TOKEN") != "" && a.getenv("ANTHROPIC_BASE_URL") != ""
}

// environ returns a.Env with IS_SANDBOX=1 and the flag that reads the CLAUDE.md of every added directory.
func (a *Agent) environ() []string {
	env := append(append([]string{}, a.Env...), "IS_SANDBOX=1")
	if a.getenv("TASK_UI_HOME") != "" {
		env = append(env, "CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD=1")
	}
	return env
}

// homeArgs returns --add-dir TASK_UI_HOME unless the home is unset or is dir itself.
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

// ask runs claude -p on prompt in dir, streams its work to a.Stdout and writes the final answer to out.
func (a *Agent) ask(ctx context.Context, dir, prompt, system string, out io.Writer) error {
	if !a.signedIn() {
		return errNoToken
	}
	path, err := a.claudePath()
	if err != nil {
		return err
	}
	args := append([]string{"-p", "--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose"}, a.homeArgs(dir)...)
	if config := a.getenv("TASK_UI_MCP_CONFIG"); config != "" {
		args = append(args, "--mcp-config", config)
	}
	if system != "" {
		args = append(args, "--append-system-prompt", system)
	}
	cmd := exec.CommandContext(ctx, path, append(args, prompt)...)
	cmd.Dir = dir
	cmd.Env = a.environ()
	cmd.Stdin, cmd.Stderr = a.Stdin, a.Stderr
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	streamClaude(pipe, a.Stdout, out)
	return cmd.Wait()
}

// claudeEvent is the part of a stream-json line the run reads: what Claude writes, does and answers.
type claudeEvent struct {
	Type    string `json:"type"`
	Message struct {
		Content []struct {
			Type  string          `json:"type"`
			Text  string          `json:"text"`
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input"`
		} `json:"content"`
	} `json:"message"`
	Result string `json:"result"`
}

// streamClaude writes each message and tool Claude writes to progress, and the run's answer to answer.
func streamClaude(r io.Reader, progress, answer io.Writer) {
	if answer == nil {
		answer = io.Discard
	}
	said := false
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64<<10), 16<<20)
	for scanner.Scan() {
		line := scanner.Bytes()
		var event claudeEvent
		if json.Unmarshal(line, &event) != nil {
			fmt.Fprintln(progress, string(line))
			continue
		}
		if event.Type == "assistant" {
			for _, block := range event.Message.Content {
				if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
					fmt.Fprintln(progress, block.Text)
					said = true
				}
				if block.Type == "tool_use" {
					fmt.Fprintln(progress, toolLine(block.Name, block.Input))
				}
			}
		}
		if event.Type == "result" && event.Result != "" {
			fmt.Fprintln(answer, event.Result)
			if !said {
				fmt.Fprintln(progress, event.Result)
			}
		}
	}
}

// toolLine names a tool Claude uses and the value that says what it works on.
func toolLine(name string, input json.RawMessage) string {
	fields := map[string]any{}
	_ = json.Unmarshal(input, &fields)
	detail := ""
	for _, key := range []string{"description", "command", "file_path", "path", "pattern", "query", "url", "prompt"} {
		if value, ok := fields[key].(string); ok && value != "" {
			detail = strings.Join(strings.Fields(value), " ")
			break
		}
	}
	if len(detail) > 160 {
		detail = detail[:160] + "..."
	}
	if detail == "" {
		return "> " + name
	}
	return "> " + name + ": " + detail
}
