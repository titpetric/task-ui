//go:build integration

package mcp_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
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

// A note naming the agent is the agent's to list; it replies and
// closes it with note_comment.
func TestTools_Notes(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	status, raw := f.post(t, f.tokens["ops"], "/api/notes", api.NoteRequest{Body: "@admin tidy the README"})
	assert.Equal(t, http.StatusCreated, status, string(raw))
	f.post(t, f.tokens["ops"], "/api/notes", api.NoteRequest{Body: "@guest have a look"})
	// A prompt from a folder page is a note too, which the board leaves out.
	assert.NoError(t, f.notes.Create(context.Background(), &model.Note{UserName: "admin", Body: "@admin a prompt", State: model.NoteOpen, WorkspaceID: model.Default, Dir: "site"}))

	text, failed := f.tool(t, "admin", "list_notes", nil)
	assert.False(t, failed, text)
	assert.Contains(t, text, "by ops, open,")
	assert.Contains(t, text, "@admin tidy the README")
	assert.NotContains(t, text, "@guest")
	assert.NotContains(t, text, "a prompt")
	id := strings.Fields(strings.TrimPrefix(text, "note "))[0]

	text, failed = f.tool(t, "admin", "note_comment", map[string]any{"id": id, "body": "Tidied in abc123.", "state": "done"})
	assert.False(t, failed, text)
	assert.Contains(t, text, "it is done")

	text, _ = f.tool(t, "admin", "list_notes", nil)
	assert.Equal(t, "No notes.", text)
	text, _ = f.tool(t, "admin", "list_notes", map[string]any{"state": "*", "mention": "*"})
	assert.Contains(t, text, "reply by admin")
	assert.Contains(t, text, "@guest have a look")
	assert.NotContains(t, text, "a prompt")

	_, failed = f.tool(t, "admin", "note_comment", map[string]any{"id": id})
	assert.True(t, failed)
	_, failed = f.tool(t, "admin", "note_comment", map[string]any{"id": "nothing", "body": "x"})
	assert.True(t, failed)
}

// A note naming a user other than the agent is that user's to list by
// default and to report on; someone else may not set its state.
func TestTools_NotesToUser(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	status, raw := f.post(t, f.tokens["ops"], "/api/notes", api.NoteRequest{Body: "@consuela tidy the README"})
	assert.Equal(t, http.StatusCreated, status, string(raw))

	text, failed := f.tool(t, "consuela", "list_notes", nil)
	assert.False(t, failed, text)
	assert.Contains(t, text, "by ops, open,")
	assert.Contains(t, text, "@consuela tidy the README")
	id := strings.Fields(strings.TrimPrefix(text, "note "))[0]

	text, failed = f.tool(t, "guest", "note_comment", map[string]any{"id": id, "state": "done"})
	assert.True(t, failed, text)
	text, failed = f.tool(t, "consuela", "note_comment", map[string]any{"id": id, "state": "working"})
	assert.False(t, failed, text)
	status, raw = f.post(t, f.tokens["ops"], "/api/notes/"+id+"/comments", map[string]string{"state": "done"})
	assert.Equal(t, http.StatusCreated, status, string(raw))
	text, failed = f.tool(t, "consuela", "note_comment", map[string]any{"id": id, "body": "Tidied in abc123.", "state": "done"})
	assert.False(t, failed, text)
	text, _ = f.tool(t, "consuela", "list_notes", nil)
	assert.Equal(t, "No notes.", text)
}
