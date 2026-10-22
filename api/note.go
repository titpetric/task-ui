package api

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// board returns the notes board, or answers 404 when notes are off.
func (h *Handlers) board(w http.ResponseWriter) bool {
	if h.opts.Board == nil {
		fail(w, http.StatusNotFound, "notes are off")
		return false
	}
	return true
}

// listNotes returns the threads of the board, newest first, each with its
// replies: ?mention= a user name, ?state= open, working, done or failed,
// ?limit=.
func (h *Handlers) listNotes(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "note/view") {
		return
	}
	if !h.board(w) {
		return
	}
	q := r.URL.Query()
	filter := storage.NoteFilter{Channel: model.ChannelNotes, Mention: q.Get("mention"), State: q.Get("state"), Limit: queryInt(r, "limit", 50)}
	h.writeThreads(w, r, filter)
}

// writeThreads answers with the threads a filter lists, each with its
// replies.
func (h *Handlers) writeThreads(w http.ResponseWriter, r *http.Request, filter storage.NoteFilter) {
	notes, err := h.opts.Board.Notes.List(r.Context(), filter)
	if err != nil {
		failErr(w, err)
		return
	}
	ids := make([]string, len(notes))
	for i, note := range notes {
		ids[i] = note.ID
	}
	replies, err := h.opts.Board.Notes.Replies(r.Context(), ids)
	if err != nil {
		failErr(w, err)
		return
	}
	views := make([]NoteView, 0, len(notes))
	for _, note := range notes {
		views = append(views, NoteView{Note: note, Replies: append(replies[note.ID][:0:0], replies[note.ID]...)})
	}
	writeJSON(w, http.StatusOK, map[string]any{"notes": views})
}

// postNote leaves a note or a reply as the user.
func (h *Handlers) postNote(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "note/edit") {
		return
	}
	if !h.board(w) {
		return
	}
	var req NoteRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	note, err := h.opts.Board.Post(r.Context(), access.UserFrom(r.Context()), req.ParentID, req.Body)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, note)
}

// editNote puts a new body in a note, its author's or any for an admin;
// an instruction or a prompt runs again, and the agent's reply to it
// changes in place. The body is a NoteRequest's, whose parent it ignores.
func (h *Handlers) editNote(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "note/edit") {
		return
	}
	if !h.board(w) {
		return
	}
	var req NoteRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	note, err := h.opts.Board.Edit(r.Context(), access.UserFrom(r.Context()), param(r, "id"), req.Body)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, note)
}

// commentNote replies to a note and, with a state, moves its instruction.
func (h *Handlers) commentNote(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "note/edit") {
		return
	}
	if !h.board(w) {
		return
	}
	var req CommentRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	note, err := h.opts.Board.Comment(r.Context(), access.UserFrom(r.Context()), param(r, "id"), req.Body, req.State)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, note)
}

// deleteNote removes a note and its replies: its author's, or any for an
// admin.
func (h *Handlers) deleteNote(w http.ResponseWriter, r *http.Request) {
	if !h.board(w) {
		return
	}
	user := access.UserFrom(r.Context())
	note, err := h.opts.Board.Notes.Get(r.Context(), param(r, "id"))
	if err != nil {
		failErr(w, err)
		return
	}
	// A prompt takes prompt/delete, any other note note/delete; either
	// way someone else's note is an admin's.
	allowed := h.can(r, "note/delete")
	if note.IsPrompt() && note.ParentID == "" {
		allowed = h.can(r, "prompt/delete")
	}
	if !allowed || (note.UserName != user.Name && !user.CanAdmin()) {
		fail(w, http.StatusForbidden, "only its author or an admin deletes a note")
		return
	}
	if err := h.opts.Board.Notes.Delete(r.Context(), note.ID); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
