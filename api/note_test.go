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

	viewer := f.call(t, asViewer, http.MethodPost, "/notes", api.NoteRequest{Body: "hello @ops"})
	assert.Equal(t, http.StatusCreated, viewer.Status)
	reply := f.call(t, asAdmin, http.MethodPost, "/notes", api.NoteRequest{ParentID: id, Body: "seconded"})
	assert.Equal(t, http.StatusCreated, reply.Status)

	board := f.call(t, asViewer, http.MethodGet, "/notes", nil)
	assert.Equal(t, http.StatusOK, board.Status)
	notes := board.list("notes")
	assert.Len(t, notes, 2)
	assert.Len(t, notes[1].(map[string]any)["replies"].([]any), 1)

	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/notes?mention=claude&state=open", nil).list("notes"), 1)
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/notes?mention=ops", nil).list("notes"), 1)

	done := f.call(t, asAdmin, http.MethodPost, "/notes/"+id+"/comments", api.CommentRequest{Body: "Done.", State: model.NoteDone})
	assert.Equal(t, http.StatusCreated, done.Status, string(done.Raw))
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/notes?state=done", nil).list("notes"), 1)
	// * is anyone and any state, as list_notes takes it.
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/notes?mention=*&state=*", nil).list("notes"), 2)
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/notes?mention=*&state=done", nil).list("notes"), 1)

	bad := f.call(t, asOps, http.MethodPost, "/notes", api.NoteRequest{ParentID: reply.str("id"), Body: "a reply to a reply"})
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

// Its author or an admin edits a note, which keeps its ID and shows
// edited.
func TestNote_Edit(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	id := f.call(t, asOps, http.MethodPost, "/notes", api.NoteRequest{Body: "a note"}).str("id")

	edited := f.call(t, asOps, http.MethodPut, "/notes/"+id, api.NoteRequest{Body: "a note, edited"})
	assert.Equal(t, http.StatusOK, edited.Status, string(edited.Raw))
	assert.Equal(t, id, edited.str("id"))
	assert.Equal(t, "a note, edited", edited.str("body"))
	assert.NotEqual(t, "", edited.str("edited_at"))
	assert.Equal(t, http.StatusForbidden, f.call(t, asViewer, http.MethodPut, "/notes/"+id, api.NoteRequest{Body: "mine"}).Status)
	assert.Equal(t, http.StatusOK, f.call(t, asAdmin, http.MethodPut, "/notes/"+id, api.NoteRequest{Body: "tidied"}).Status)
	assert.Equal(t, http.StatusBadRequest, f.call(t, asOps, http.MethodPut, "/notes/"+id, api.NoteRequest{Body: " "}).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asOps, http.MethodPut, "/notes/nothing", api.NoteRequest{Body: "x"}).Status)
}
