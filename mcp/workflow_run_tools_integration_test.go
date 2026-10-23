//go:build integration

package mcp_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// The tools list a workflow, run it by name, wait for it and list its runs.
func TestTools_Workflows(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	status, raw := f.post(t, f.tokens["admin"], "/api/workflows", api.WorkflowRequest{Name: "Ship", Steps: []api.StepRequest{{Folder: "site", Job: "build"}}})
	assert.Equal(t, http.StatusCreated, status, string(raw))

	text, _ := f.tool(t, "admin", "list_workflows", nil)
	assert.Contains(t, text, "Ship (id ")
	assert.Contains(t, text, "): site/build")

	text, failed := f.tool(t, "admin", "run_workflow", map[string]any{"workflow": "ship", "wait": 20})
	assert.False(t, failed, text)
	assert.Contains(t, text, "Ship by admin, ok")
	assert.Contains(t, text, "1. site/build: ok (run ")

	text, _ = f.tool(t, "admin", "list_workflow_runs", map[string]any{"workflow": "Ship"})
	assert.Contains(t, text, "Ship by admin, ok")

	_, failed = f.tool(t, "admin", "run_workflow", map[string]any{"workflow": "nothing"})
	assert.True(t, failed)
	_, failed = f.tool(t, "guest", "run_workflow", map[string]any{"workflow": "Ship"})
	assert.True(t, failed)
}

// run_workflow passes its input to the workflow's prompt, after it
// under Input:.
func TestTools_RunWorkflowInput(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.board.AgentCommand = []string{"sh", "-c", `printf 'Asked: %s' "$TASK_UI_PROMPT" > "$TASK_UI_PROMPT_REPLY"`}
	status, raw := f.post(t, f.tokens["admin"], "/api/workflows", api.WorkflowRequest{Name: "Asks", Prompt: "say ok, change nothing"})
	assert.Equal(t, http.StatusCreated, status, string(raw))

	text, failed := f.tool(t, "admin", "run_workflow", map[string]any{"workflow": "Asks", "input": "loudly", "wait": 20})
	assert.False(t, failed, text)
	assert.Contains(t, text, "Asks by admin, ok")
	threads, err := f.notes.List(context.Background(), storage.NoteFilter{Prompts: true})
	assert.NoError(t, err)
	assert.Len(t, threads, 1)
	assert.Equal(t, "say ok, change nothing\n\nInput:\nloudly", threads[0].Body)
	assert.True(t, threads[0].AboutWorkflowRunID != "")
}
