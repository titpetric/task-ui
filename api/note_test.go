package api_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A note to the agent opens an instruction; the board lists it with its
// replies, filtered by mention and state, and the agent reports on it.
func TestNote_PostListComment(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.call(t, asOps, http.MethodPost, "/notes", api.NoteRequest{Body: "@claude tidy the README"})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	id := res.str("id")
	assert.Equal(t, model.NoteOpen, res.str("state"))

	chat := f.call(t, asViewer, http.MethodPost, "/notes", api.NoteRequest{Channel: model.ChannelChat, Body: "hello @ops"})
	assert.Equal(t, http.StatusCreated, chat.Status)
	reply := f.call(t, asAdmin, http.MethodPost, "/notes", api.NoteRequest{ParentID: id, Body: "seconded"})
	assert.Equal(t, http.StatusCreated, reply.Status)

	board := f.call(t, asViewer, http.MethodGet, "/notes?channel=notes", nil)
	assert.Equal(t, http.StatusOK, board.Status)
	notes := board.list("notes")
	assert.Len(t, notes, 1)
	assert.Len(t, notes[0].(map[string]any)["replies"].([]any), 1)

	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/notes?mention=claude&state=open", nil).list("notes"), 1)
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/notes?mention=ops", nil).list("notes"), 1)

	done := f.call(t, asAdmin, http.MethodPost, "/notes/"+id+"/comments", api.CommentRequest{Body: "Done.", State: model.NoteDone})
	assert.Equal(t, http.StatusCreated, done.Status, string(done.Raw))
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/notes?state=done", nil).list("notes"), 1)

	bad := f.call(t, asOps, http.MethodPost, "/notes", api.NoteRequest{Channel: "nowhere", Body: "x"})
	assert.Equal(t, http.StatusBadRequest, bad.Status)
	empty := f.call(t, asOps, http.MethodPost, "/notes", api.NoteRequest{Body: " "})
	assert.Equal(t, http.StatusBadRequest, empty.Status)
	missing := f.call(t, asOps, http.MethodPost, "/notes/nothing/comments", api.CommentRequest{Body: "x"})
	assert.Equal(t, http.StatusNotFound, missing.Status)
}

// Only its author or an admin deletes a note.
func TestNote_Delete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	id := f.call(t, asOps, http.MethodPost, "/notes", api.NoteRequest{Body: "a note"}).str("id")

	assert.Equal(t, http.StatusForbidden, f.call(t, asViewer, http.MethodDelete, "/notes/"+id, nil).Status)
	assert.Equal(t, http.StatusNoContent, f.call(t, asOps, http.MethodDelete, "/notes/"+id, nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodDelete, "/notes/"+id, nil).Status)
}
