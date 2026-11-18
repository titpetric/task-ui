package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/feature/note"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// heartbeat is how often an idle stream sends a comment.
const heartbeat = 15 * time.Second

// threadVisible reports whether the request's user may follow a thread.
func (h *Handlers) threadVisible(r *http.Request, thread *model.Note) bool {
	user := access.UserFrom(r.Context())
	switch {
	case thread.IsPrompt():
		return h.can(r, "prompt/view") && seesPlace(h.opts.Catalogue, user, thread.Dir)
	case thread.Channel == model.ChannelChat:
		if !h.can(r, "note/view") || h.opts.Board.Chats == nil {
			return false
		}
		member, err := h.opts.Board.Chats.IsParticipant(r.Context(), thread.ID, user.ID)
		return err == nil && (member || user.CanAdmin())
	default:
		return h.can(r, "note/view")
	}
}

// sse streams the events of a topic as server-sent events.
func (h *Handlers) sse(w http.ResponseWriter, r *http.Request) {
	topic := param(r, "topic")
	kind, id, _ := strings.Cut(topic, ":")
	user := access.UserFrom(r.Context())
	// visible says whether the user may see an event of the topic.
	visible := func(event.Event) bool { return true }
	switch kind {
	case runner.RunsTopic:
		if id != "" {
			failErr(w, storage.ErrNotFound)
			return
		}
		prompts, notes := h.can(r, "prompt/view"), h.can(r, "note/view")
		visible = func(e event.Event) bool {
			run, ok := e.Data.(model.Run)
			return ok && user.SeesRun(&run, prompts, notes)
		}
	case note.Topic:
		if id != "" || h.opts.Board == nil {
			failErr(w, storage.ErrNotFound)
			return
		}
		notes, prompts := h.can(r, "note/view"), h.can(r, "prompt/view")
		visible = func(e event.Event) bool {
			n, ok := e.Data.(model.Note)
			if !ok {
				return false
			}
			if n.IsPrompt() {
				return prompts
			}
			if n.Channel == model.ChannelChat {
				if h.opts.Board.Chats == nil {
					return false
				}
				root := n.ID
				if n.ParentID != "" {
					root = n.ParentID
				}
				member, err := h.opts.Board.Chats.IsParticipant(r.Context(), root, user.ID)
				return notes && err == nil && (member || user.CanAdmin())
			}
			return notes
		}
	case "thread":
		if id == "" || h.opts.Board == nil {
			failErr(w, storage.ErrNotFound)
			return
		}
		thread, err := h.opts.Board.Notes.Get(r.Context(), id)
		if err != nil || thread.ParentID != "" || !h.threadVisible(r, thread) {
			failErr(w, storage.ErrNotFound)
			return
		}
	case "workflow":
		wf, err := h.opts.Workflows.Get(r.Context(), id)
		if err != nil {
			failErr(w, err)
			return
		}
		steps, err := h.opts.Workflows.Steps(r.Context(), wf.ID)
		if err != nil || !h.visibleSteps(r, steps) {
			failErr(w, storage.ErrNotFound)
			return
		}
	default:
		failErr(w, storage.ErrNotFound)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok || h.opts.Events == nil {
		fail(w, http.StatusNotImplemented, "live events aren't available on this instance")
		return
	}

	events, cancel := h.opts.Events.Subscribe(topic)
	defer cancel()
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case e, open := <-events:
			if !open {
				return
			}
			if !visible(e) {
				continue
			}
			data, err := json.Marshal(e.Data)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", e.Name, data)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}
