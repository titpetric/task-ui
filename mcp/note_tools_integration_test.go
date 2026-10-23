//go:build integration

package mcp_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// post_note leaves a note as the token's user, or a reply; it takes
// note/edit, which a user policy can deny.
func TestNoteTools_PostNote(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	text, failed := f.tool(t, "ops", "post_note", map[string]any{"body": "@admin look at the build"})
	assert.False(t, failed, text)
	assert.Contains(t, text, "as ops.")
	id := strings.Fields(strings.TrimPrefix(text, "Left note "))[0]

	text, failed = f.tool(t, "guest", "post_note", map[string]any{"body": "seen", "parent": id})
	assert.False(t, failed, text)
	assert.Contains(t, text, "Replied to note "+id+" as guest")

	text, _ = f.tool(t, "admin", "list_notes", nil)
	assert.Contains(t, text, "@admin look at the build")
	assert.Contains(t, text, "reply by guest")

	_, failed = f.tool(t, "ops", "post_note", map[string]any{"body": " "})
	assert.True(t, failed)
	status, raw := f.put(t, f.tokens["admin"], "/api/permissions/users/ops", map[string]string{"note/edit": "deny"})
	assert.Equal(t, 200, status, string(raw))
	text, failed = f.tool(t, "ops", "post_note", map[string]any{"body": "again"})
	assert.True(t, failed)
	assert.Contains(t, text, "(403)")
}

// ask_prompt takes prompt/edit, a folder and a prompt.
func TestNoteTools_AskPrompt(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	text, failed := f.tool(t, "ops", "ask_prompt", map[string]any{"folder": "site", "prompt": "what is here"})
	assert.True(t, failed)
	assert.Contains(t, text, "(403)")
	_, failed = f.tool(t, "admin", "ask_prompt", map[string]any{"folder": "site"})
	assert.True(t, failed)
	text, failed = f.tool(t, "admin", "ask_prompt", map[string]any{"folder": "nothing", "prompt": "what is here"})
	assert.True(t, failed)
	assert.Contains(t, text, "(404)")
}
