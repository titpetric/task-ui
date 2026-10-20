package mcp_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// A job is listed, run, waited for and shown; a viewer may not run it.
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
}

// A workflow is listed, run by name and waited for, and its runs listed.
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
