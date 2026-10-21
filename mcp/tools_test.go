package mcp_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// The tools list a job, run it, wait for it and show it; a viewer may not run it.
func TestTools_Jobs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	text, failed := f.tool(t, "admin", "list_jobs", nil)
	assert.False(t, failed)
	assert.Contains(t, text, "site\n  build - Build the site")

	text, failed = f.tool(t, "guest", "run_job", map[string]any{"folder": "site", "job": "build"})
	assert.True(t, failed)
	assert.Contains(t, text, "not run it (403)")

	text, failed = f.tool(t, "admin", "run_job", map[string]any{"folder": "site", "job": "build", "wait": 20})
	assert.False(t, failed, text)
	assert.Contains(t, text, "site/build by admin, ok (exit 0)")
	assert.Contains(t, text, "built")
	id := strings.Fields(strings.TrimPrefix(text, "run "))[0]
	id = strings.TrimSuffix(id, ":")

	text, _ = f.tool(t, "guest", "list_jobs", map[string]any{"folder": "site"})
	assert.Contains(t, text, "[view only] | last: ok (exit 0)")

	text, _ = f.tool(t, "guest", "list_runs", map[string]any{"folder": "site", "limit": 5})
	assert.Contains(t, text, id+" site/build by admin: ok")

	text, failed = f.tool(t, "guest", "get_run", map[string]any{"id": id, "lines": 3})
	assert.False(t, failed, text)
	assert.Contains(t, text, "run "+id)

	_, failed = f.tool(t, "admin", "get_run", map[string]any{"id": "nothing"})
	assert.True(t, failed)

	// An interactive shell is an admin's alone.
	for _, user := range []string{"ops", "guest"} {
		text, failed = f.tool(t, user, "run_job", map[string]any{"folder": "site", "job": "@shell"})
		assert.True(t, failed, user)
		assert.Contains(t, text, "(403)", user)
	}
}

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

// A note naming the agent is the agent's to list; it replies and
// closes it with note_comment.
func TestTools_Notes(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	status, raw := f.post(t, f.tokens["ops"], "/api/notes", api.NoteRequest{Body: "@admin tidy the README"})
	assert.Equal(t, http.StatusCreated, status, string(raw))
	f.post(t, f.tokens["ops"], "/api/notes", api.NoteRequest{Body: "@guest have a look"})

	text, failed := f.tool(t, "admin", "list_notes", nil)
	assert.False(t, failed, text)
	assert.Contains(t, text, "by ops, open,")
	assert.Contains(t, text, "@admin tidy the README")
	assert.NotContains(t, text, "@guest")
	id := strings.Fields(strings.TrimPrefix(text, "note "))[0]

	text, failed = f.tool(t, "admin", "note_comment", map[string]any{"id": id, "body": "Tidied in abc123.", "state": "done"})
	assert.False(t, failed, text)
	assert.Contains(t, text, "it is done")

	text, _ = f.tool(t, "admin", "list_notes", nil)
	assert.Equal(t, "No notes.", text)
	text, _ = f.tool(t, "admin", "list_notes", map[string]any{"state": "*", "mention": "*"})
	assert.Contains(t, text, "reply by admin")
	assert.Contains(t, text, "@guest have a look")

	_, failed = f.tool(t, "admin", "note_comment", map[string]any{"id": id})
	assert.True(t, failed)
	_, failed = f.tool(t, "admin", "note_comment", map[string]any{"id": "nothing", "body": "x"})
	assert.True(t, failed)
}
