package mcp_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// whoami says what the token's user may do, by role and own policy, and
// a refused call names the permission it needs and points there.
func TestTools_whoami(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	text, failed := f.tool(t, "ops", "whoami", nil)
	assert.False(t, failed, text)
	assert.Contains(t, text, "You are ops, role runner.")
	assert.Contains(t, text, "everything else is denied")
	assert.Contains(t, text, "  Run Jobs (run): allowed.")
	assert.Contains(t, text, "  Secrets (secret): denied. Variables added to the environment of runs")
	assert.Contains(t, text, "  Prompts (prompt): denied.")
	assert.Contains(t, text, "Folders you see (1): site")
	assert.Contains(t, text, "taskui://context")

	text, _ = f.tool(t, "guest", "whoami", nil)
	assert.Contains(t, text, "You are guest, role viewer.")
	assert.Contains(t, text, "  Run Jobs (run): denied.")

	// A policy of the user's own, as the claude user holds.
	status, raw := f.put(t, f.tokens["admin"], "/api/permissions/users/consuela", map[string]string{"prompt/view": "allow", "prompt/edit": "allow", "note/delete": "deny"})
	assert.Equal(t, http.StatusOK, status, string(raw))
	text, _ = f.tool(t, "consuela", "whoami", nil)
	assert.Contains(t, text, "  Prompts (prompt): view, edit.")
	assert.Contains(t, text, "prompt/view, prompt/edit")
	assert.NotContains(t, text, "note/delete")

	text, _ = f.tool(t, "admin", "whoami", nil)
	assert.Contains(t, text, "You may: everything, as an admin.")

	// A refusal names the permission and whoami.
	text, failed = f.tool(t, "ops", "ask_prompt", map[string]any{"folder": "site", "prompt": "what builds here"})
	assert.True(t, failed)
	assert.Contains(t, text, "needs prompt/edit (403); call whoami to see what you may do")
}
