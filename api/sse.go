package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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
//
// The stream ends when the client goes.
func (h *Handlers) sse(w http.ResponseWriter, r *http.Request) {
	topic := param(r, "topic")
	kind, id, _ := strings.Cut(topic, ":")
	switch kind {
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
