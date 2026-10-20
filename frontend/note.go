package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// notes shows the notes board: the notes newest first, each with its
// replies, and the agent a note names to start work.
func (h *Handlers) notes(w http.ResponseWriter, r *http.Request) {
	if h.opts.Board == nil {
		h.missing(w, r, "Notes are off.")
		return
	}
	p := h.page(r, "Notes")
	p.Crumbs = view.Crumbs("Notes")
	p.Agent = h.opts.Board.Agent
	p.Threads = h.threads(r, model.ChannelNotes, 100)
	h.render(w, http.StatusOK, "notes", p)
}

// chat shows the chat: the latest messages, oldest first, as they read.
func (h *Handlers) chat(w http.ResponseWriter, r *http.Request) {
	if h.opts.Board == nil {
		h.missing(w, r, "The chat is off.")
		return
	}
	p := h.page(r, "Chat")
	p.Crumbs = view.Crumbs("Chat")
	threads := h.threads(r, model.ChannelChat, 200)
	for i, j := 0, len(threads)-1; i < j; i, j = i+1, j-1 {
		threads[i], threads[j] = threads[j], threads[i]
	}
	p.Threads = threads
	h.render(w, http.StatusOK, "chat", p)
}

// threads lists a channel's notes, newest first, with their replies.
func (h *Handlers) threads(r *http.Request, channel string, limit int) []view.Thread {
	notes, _ := h.opts.Board.Notes.List(r.Context(), storage.NoteFilter{Channel: channel, Limit: limit})
	ids := make([]string, len(notes))
	for i, note := range notes {
		ids[i] = note.ID
	}
	replies, _ := h.opts.Board.Notes.Replies(r.Context(), ids)
	threads := make([]view.Thread, 0, len(notes))
	for _, note := range notes {
		threads = append(threads, view.Thread{Note: note, Replies: replies[note.ID]})
	}
	return threads
}
