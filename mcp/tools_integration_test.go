//go:build integration

package mcp_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// The tools list a job, run it, wait for it and show it; a viewer may not run it.
func TestTools_Jobs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	text, failed := f.tool(t, "admin", "list_jobs", nil)
	assert.False(t, failed)
	assert.Contains(t, text, "site\n  default - Check the site\n  build - Build the site", "default first")

	text, failed = f.tool(t, "guest", "run_job", map[string]any{"folder": "site", "job": "build"})
	assert.True(t, failed)
	assert.Contains(t, text, "permission to run it (403)")

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
