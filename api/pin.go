package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// pinUser returns the signed in user of a request, answering 404 with
// pins off and 401 without a user; nil then.
func (h *Handlers) pinUser(w http.ResponseWriter, r *http.Request) *model.User {
	user := access.UserFrom(r.Context())
	switch {
	case h.opts.Pins == nil:
		fail(w, http.StatusNotFound, "pins are off")
		return nil
	case user == nil:
		fail(w, http.StatusUnauthorized, "sign in to pin")
		return nil
	}
	return user
}

// listPins returns the pins of the user asking, oldest first.
func (h *Handlers) listPins(w http.ResponseWriter, r *http.Request) {
	user := h.pinUser(w, r)
	if user == nil {
		return
	}
	pins, err := h.opts.Pins.List(r.Context(), user.Name)
	if err != nil {
		failErr(w, err)
		return
	}
	views := []PinView{}
	for _, pin := range pins {
		views = append(views, NewPinView(pin))
	}
	writeJSON(w, http.StatusOK, map[string]any{"pins": views})
}

// addPin pins a folder the user sees, a directory inside one, or a job
// of it, and answers the pin; pinning the same again answers the pin
// there is.
func (h *Handlers) addPin(w http.ResponseWriter, r *http.Request) {
	user := h.pinUser(w, r)
	if user == nil {
		return
	}
	var req PinRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Folder = strings.Trim(req.Folder, "/")
	if problem := h.pinTarget(r, req); problem != "" {
		fail(w, http.StatusNotFound, problem)
		return
	}
	if len(req.Args) > 0 {
		if _, job, _, _ := h.targetIn(r, req.Folder, req.Dir, req.Job); !job.TakesArgs {
			fail(w, http.StatusBadRequest, "the job takes no args")
			return
		}
	}
	args, _ := json.Marshal([]string(req.Args))
	if len(req.Args) == 0 {
		args = []byte("[]")
	}
	pin := &model.Pin{UserName: user.Name, Folder: req.Folder, Job: req.Job, Dir: req.Dir, Args: string(args)}
	if err := h.opts.Pins.Add(r.Context(), pin); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, NewPinView(*pin))
}

// pinTarget says what a pin request names that the user does not see
// or that is not there, "" for a target there: a job of a folder, or a folder or
// a directory inside one; a dir or args need a job.
func (h *Handlers) pinTarget(r *http.Request, req PinRequest) string {
	if req.Job != "" {
		_, _, _, problem := h.targetIn(r, req.Folder, req.Dir, req.Job)
		return problem
	}
	if req.Dir != "" || len(req.Args) > 0 {
		return "a dir or args pin a job: name it"
	}
	f, rest := folder.Resolve(h.every(), req.Folder)
	if f == nil || !access.UserFrom(r.Context()).CanSee(f.ID) {
		return "no such folder"
	}
	if rest == "" {
		return ""
	}
	dir, ok := f.Path(rest)
	if info, err := os.Stat(dir); !ok || err != nil || !info.IsDir() {
		return "no such folder"
	}
	return ""
}

// deletePin removes a pin of the user asking.
func (h *Handlers) deletePin(w http.ResponseWriter, r *http.Request) {
	user := h.pinUser(w, r)
	if user == nil {
		return
	}
	if err := h.opts.Pins.Delete(r.Context(), user.Name, param(r, "id")); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
