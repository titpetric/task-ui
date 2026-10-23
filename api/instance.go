package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/runner"
)

// InstanceRequest sets the instance settings the Instance page edits.
type InstanceRequest struct {
	// BacklogLength is how many runs may wait for a free slot, as a
	// number or a string; empty or 0 is the automatic length,
	// backlog_length_auto.
	BacklogLength *backlogValue `json:"backlog_length"`
}

// backlogValue takes a backlog length as a JSON number or string.
type backlogValue string

// UnmarshalJSON reads a number, a string or null.
func (v *backlogValue) UnmarshalJSON(data []byte) error {
	text := string(data)
	switch {
	case text == "null":
		*v = ""
	case len(text) > 1 && text[0] == '"':
		unquoted, err := strconv.Unquote(text)
		if err != nil {
			return err
		}
		*v = backlogValue(unquoted)
	default:
		*v = backlogValue(text)
	}
	return nil
}

// getInstance answers how the slots and the backlog stand: the runs
// that go at once, the ones that wait, and the backlog lengths, the
// automatic one, the one set and the one in effect.
func (h *Handlers) getInstance(w http.ResponseWriter, r *http.Request) {
	if h.opts.Runner == nil {
		fail(w, http.StatusNotFound, "the runner isn't available on this instance")
		return
	}
	writeJSON(w, http.StatusOK, h.opts.Runner.Backlog(r.Context()))
}

// putInstance sets the backlog length, for admins alone; the next start
// reads it.
func (h *Handlers) putInstance(w http.ResponseWriter, r *http.Request) {
	if !access.UserFrom(r.Context()).CanAdmin() {
		fail(w, http.StatusForbidden, "only an admin sets the instance settings")
		return
	}
	if h.opts.Runner == nil {
		fail(w, http.StatusNotFound, "the runner isn't available on this instance")
		return
	}
	var req InstanceRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.BacklogLength != nil {
		if err := h.opts.Runner.SetBacklog(r.Context(), string(*req.BacklogLength)); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, runner.ErrBacklogInvalid) {
				status = http.StatusBadRequest
			}
			fail(w, status, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, h.opts.Runner.Backlog(r.Context()))
}
