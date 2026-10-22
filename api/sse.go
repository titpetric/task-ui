package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// heartbeat is how often an idle stream sends a comment, so proxies and
// browsers keep it open.
const heartbeat = 15 * time.Second

// sse streams the events of a topic as server-sent events: `event:` is
// the name, `data:` the JSON payload. The topics:
//
//	workflow:{id}  a workflow's runs: run, step and done, for a user who
//	               sees every step of it
//	runs           every run as it starts and ends: run, with the run row,
//	               of the folders the user sees; a prompt's with prompt/view
//	notes          every note and reply as someone leaves or changes it:
//	               note, with note/view; a prompt's notes with prompt/view
//
// The stream ends when the client goes.
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
			if n.IsPrompt() {
				return ok && prompts
			}
			return ok && notes
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
		fail(w, http.StatusNotImplemented, "events are not available")
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
