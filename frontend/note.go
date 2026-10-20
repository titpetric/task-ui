package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
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

// threads lists a channel's notes, newest first, with their replies. A
// note links its agent's run only for a user who sees the run.
func (h *Handlers) threads(r *http.Request, channel string, limit int) []view.Thread {
	notes, _ := h.opts.Board.Notes.List(r.Context(), storage.NoteFilter{Channel: channel, Limit: limit})
	ids := make([]string, len(notes))
	for i, note := range notes {
		ids[i] = note.ID
	}
	replies, _ := h.opts.Board.Notes.Replies(r.Context(), ids)
	sees := h.seesRun(r)
	threads := make([]view.Thread, 0, len(notes))
	for _, note := range notes {
		thread := view.Thread{Note: note, Replies: replies[note.ID]}
		if !sees(thread.RunID) {
			thread.RunID = ""
		}
		for i := range thread.Replies {
			if !sees(thread.Replies[i].RunID) {
				thread.Replies[i].RunID = ""
			}
		}
		threads = append(threads, thread)
	}
	return threads
}

// seesRun returns a check of whether the user sees a run, by its folder,
// that looks each run up once. A run that is not there links nowhere.
func (h *Handlers) seesRun(r *http.Request) func(id string) bool {
	user := access.UserFrom(r.Context())
	seen := map[string]bool{}
	return func(id string) bool {
		if id == "" {
			return false
		}
		if sees, ok := seen[id]; ok {
			return sees
		}
		run, err := h.opts.Runs.Get(r.Context(), id)
		seen[id] = err == nil && user.CanSee(run.Folder)
		return seen[id]
	}
}
