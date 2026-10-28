package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/toolenv"
)

// InstanceRequest sets the instance settings the Instance page edits.
type InstanceRequest struct {
	// BacklogLength is how many runs may wait for a free slot, as a
	// number or a string; empty or 0 is the automatic length,
	// backlog_length_auto.
	BacklogLength *backlogValue `json:"backlog_length"`
	// ParentEnv is how much of the server's environment a run starts
	// with: none, auto or all.
	ParentEnv *string `json:"parent_env"`
}

// toolRow is a detected tool as the tool table keeps it.
func toolRow(tool toolenv.Tool) model.Tool {
	return model.Tool{Name: tool.Name, Category: tool.Category, Version: tool.Version, Path: tool.Path, Description: tool.Description}
}

// checkTools probes the PATH for every tool the instance knows, replaces
// the stored rows with what it found, and answers them: the Check all
// button of the Instance page. For admins alone.
func (h *Handlers) checkTools(w http.ResponseWriter, r *http.Request) {
	if !access.UserFrom(r.Context()).CanAdmin() {
		fail(w, http.StatusForbidden, "only an admin checks the instance's tooling")
		return
	}
	if h.opts.Tools == nil {
		fail(w, http.StatusNotFound, "the tool table isn't available on this instance")
		return
	}
	found := toolenv.Detect{}.Tools(r.Context())
	rows := make([]model.Tool, 0, len(found))
	for _, tool := range found {
		rows = append(rows, toolRow(tool))
	}
	if err := h.opts.Tools.Replace(r.Context(), rows); err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tools": rows})
}

// checkTool probes the PATH for one tool and stores what it found, or
// removes the row when the PATH no longer has the tool: the Check button
// beside a tool of the Instance page. For admins alone.
func (h *Handlers) checkTool(w http.ResponseWriter, r *http.Request) {
	if !access.UserFrom(r.Context()).CanAdmin() {
		fail(w, http.StatusForbidden, "only an admin checks the instance's tooling")
		return
	}
	if h.opts.Tools == nil {
		fail(w, http.StatusNotFound, "the tool table isn't available on this instance")
		return
	}
	spec, ok := toolenv.SpecByID(chi.URLParam(r, "id"))
	if !ok {
		fail(w, http.StatusNotFound, "the instance doesn't know this tool")
		return
	}
	found := toolenv.Detect{Specs: []toolenv.Spec{spec}}.Tools(r.Context())
	if len(found) == 0 {
		if err := h.opts.Tools.Remove(r.Context(), spec.Name); err != nil {
			fail(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"tool": spec.Name, "found": false})
		return
	}
	row := toolRow(found[0])
	if err := h.opts.Tools.Save(r.Context(), row); err != nil {
		fail(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tool": spec.Name, "found": true, "version": row.Version})
}

// instanceState reads the state GET and PUT /api/instance answer.
func (h *Handlers) instanceState(r *http.Request) InstanceState {
	return InstanceState{BacklogState: h.opts.Runner.Backlog(r.Context()), ParentEnv: h.opts.Runner.ParentEnv(r.Context())}
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

// getInstance answers the slots, the backlog and its lengths, and the
// parent environment, as the Instance page shows them: for admins
// alone, whose settings they are.
func (h *Handlers) getInstance(w http.ResponseWriter, r *http.Request) {
	if !access.UserFrom(r.Context()).CanAdmin() {
		fail(w, http.StatusForbidden, "only an admin reads the instance settings")
		return
	}
	if h.opts.Runner == nil {
		fail(w, http.StatusNotFound, "the runner isn't available on this instance")
		return
	}
	writeJSON(w, http.StatusOK, h.instanceState(r))
}

// putInstance sets the backlog length and the parent environment, for
// admins alone; the next start reads them.
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
	if req.ParentEnv != nil {
		if err := h.opts.Runner.SetParentEnv(r.Context(), *req.ParentEnv); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, runner.ErrEnvironmentInvalid) {
				status = http.StatusBadRequest
			}
			fail(w, status, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, h.instanceState(r))
}
