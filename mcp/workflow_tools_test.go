package mcp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// create_workflow makes a workflow of a job and a prompt step,
// get_workflow shows it, update_workflow renames it and keeps its steps,
// and run_workflow runs it and ends with the final prompt's answer. A
// prompt step takes prompt/edit.
func TestTools_WorkflowEdits(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.board.AgentCommand = []string{"sh", "-c", `printf 'Asked: %s' "$TASK_UI_PROMPT" > "$TASK_UI_PROMPT_REPLY"`}
	steps := []map[string]any{{"folder": "site", "job": "build"}, {"prompt": "fetch the weather", "folder": "_"}}

	text, failed := f.tool(t, "admin", "create_workflow", map[string]any{"name": "Weather", "steps": steps, "prompt": "water today?"})
	assert.False(t, failed, text)
	assert.Contains(t, text, "workflow Weather (id ")
	assert.Contains(t, text, "  1. site/build\n  2. prompt in _: fetch the weather\n")
	assert.Contains(t, text, "final prompt, in _: water today?")

	text, failed = f.tool(t, "ops", "create_workflow", map[string]any{"name": "Asks", "steps": steps[1:]})
	assert.True(t, failed, text)
	assert.Contains(t, text, "403")

	text, failed = f.tool(t, "admin", "update_workflow", map[string]any{"workflow": "Weather", "name": "Watering", "sidebar": false})
	assert.False(t, failed, text)
	assert.Contains(t, text, "workflow Watering (id ")
	assert.Contains(t, text, ", not in the sidebar")
	assert.Contains(t, text, "  2. prompt in _: fetch the weather\n")
	assert.Contains(t, text, "final prompt, in _: water today?")

	text, failed = f.tool(t, "admin", "get_workflow", map[string]any{"workflow": "watering"})
	assert.False(t, failed, text)
	assert.Contains(t, text, "  1. site/build\n")

	text, failed = f.tool(t, "admin", "run_workflow", map[string]any{"workflow": "Watering", "wait": 30})
	assert.False(t, failed, text)
	assert.Contains(t, text, "Watering by admin, ok")
	assert.Contains(t, text, "  2. prompt in _: ok")
	assert.Contains(t, text, "answer:\nAsked: water today?")

	_, failed = f.tool(t, "admin", "get_workflow", map[string]any{"workflow": "nothing"})
	assert.True(t, failed)
}

// A prompt's run gets an MCP configuration with a token of the asking
// user, which signs in to the MCP server while the run goes and not
// after; once the run ended, the board removed its directory.
func TestTools_RunToken(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.board.MCPURL = func() string { return f.server.URL + "/api/mcp" }
	out := t.TempDir()
	f.board.AgentCommand = []string{"sh", "-c", `echo "$TASK_UI_MCP_CONFIG" > ` + out + `/path; cp "$TASK_UI_MCP_CONFIG" ` + out + `/config.json; ` +
		`i=0; while [ ! -f ` + out + `/go ] && [ $i -lt 600 ]; do sleep 0.05; i=$((i+1)); done; echo done > "$TASK_UI_PROMPT_REPLY"`}

	text, failed := f.tool(t, "admin", "ask_prompt", map[string]any{"folder": "site", "prompt": "list the jobs"})
	assert.False(t, failed, text)
	config := filepath.Join(out, "config.json")
	assert.Eventually(t, func() bool { _, err := os.Stat(config); return err == nil }, 20*time.Second, 20*time.Millisecond)
	data, err := os.ReadFile(config)
	assert.NoError(t, err)
	var parsed struct {
		MCPServers map[string]struct {
			Type    string            `json:"type"`
			URL     string            `json:"url"`
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	assert.NoError(t, json.Unmarshal(data, &parsed))
	server := parsed.MCPServers["task"]
	assert.Equal(t, "http", server.Type)
	assert.Equal(t, f.server.URL+"/api/mcp", server.URL)
	token := strings.TrimPrefix(server.Headers["Authorization"], "Bearer ")
	assert.True(t, strings.HasPrefix(token, "tui_"))
	path, _ := os.ReadFile(filepath.Join(out, "path"))
	info, err := os.Stat(filepath.Dir(strings.TrimSpace(string(path))))
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())

	// During the run the token is the asker's, with the asker's rights,
	// and needs no MCP agent user of the asker's.
	ctx := context.Background()
	admin, err := f.board.Users.GetByName(ctx, "admin")
	assert.NoError(t, err)
	admin.Mcp = false
	assert.NoError(t, f.board.Users.Update(ctx, admin))
	status, _ := f.post(t, f.tokens["admin"], "/api/mcp", map[string]any{"jsonrpc": "2.0", "id": 1, "method": "ping"})
	assert.Equal(t, http.StatusForbidden, status)
	f.tokens["run"] = token
	text, failed = f.tool(t, "run", "whoami", nil)
	assert.False(t, failed, text)
	assert.Contains(t, text, "admin")

	assert.NoError(t, os.WriteFile(filepath.Join(out, "go"), nil, 0o644))
	assert.Eventually(t, func() bool {
		threads, _ := f.notes.List(ctx, storage.NoteFilter{Prompts: true, Dir: "site"})
		return len(threads) == 1 && threads[0].State == "done"
	}, 20*time.Second, 20*time.Millisecond)
	// The cleanup follows the answer, a moment after it.
	assert.Eventually(t, func() bool {
		_, err := os.Stat(strings.TrimSpace(string(path)))
		return os.IsNotExist(err)
	}, 10*time.Second, 20*time.Millisecond)
	status, _ = f.post(t, token, "/api/mcp", map[string]any{"jsonrpc": "2.0", "id": 1, "method": "ping"})
	assert.Equal(t, http.StatusUnauthorized, status)
}

// ask_prompt with area workflows asks as the Workflows page's prompt:
// in the root, with the builder's instruction and the run's MCP
// configuration, and waits for Claude's reply.
func TestTools_AskWorkflows(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.board.MCPURL = func() string { return f.server.URL + "/api/mcp" }
	f.board.AgentCommand = []string{"sh", "-c", `{ printf '%s\n' "$TASK_UI_PROMPT_INSTRUCTIONS"; test -f "$TASK_UI_MCP_CONFIG" && echo mcp; } > "$TASK_UI_PROMPT_REPLY"`}
	text, failed := f.tool(t, "admin", "ask_prompt", map[string]any{"area": "workflows", "prompt": "build a weather workflow", "wait": 30})
	assert.False(t, failed, text)
	assert.Contains(t, text, " in _ by admin, done")
	assert.Contains(t, text, "create_workflow or update_workflow")
	assert.Contains(t, text, "\nmcp\n")
	threads, _ := f.notes.List(context.Background(), storage.NoteFilter{Prompts: true, Area: "workflows"})
	assert.Len(t, threads, 1)
}
