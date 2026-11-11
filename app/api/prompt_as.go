package api

import (
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
)

// promptUser returns the user a prompt in where runs as, the thread's
// and the run's user: the user the request names, picked as pickRunAs
// does, who holds prompt/edit and sees where, 403 naming them otherwise;
// the caller with no name. A reply keeps its thread's user, the caller
// for a user since deleted, and takes no name (400); only an admin or
// the thread's user replies (403).
func (h *Handlers) promptUser(r *http.Request, req PromptRequest, where string) (*model.User, int, string) {
	if req.Parent != "" {
		return h.threadUser(r, req)
	}
	user, status, problem := h.pickRunAs(r, req.User)
	if problem != "" {
		return nil, status, problem
	}
	caller := access.UserFrom(r.Context())
	if user == nil || user.ID == caller.ID {
		return caller, 0, ""
	}
	if !h.opts.Users.Permissions(r.Context(), user).Allowed("prompt/edit") || !seesPlace(h.opts.Catalogue, user, where) {
		return nil, http.StatusForbidden, user.Name + " may not prompt in " + where
	}
	return user, 0, ""
}

// threadUser returns the user of the thread a reply continues.
func (h *Handlers) threadUser(r *http.Request, req PromptRequest) (*model.User, int, string) {
	if strings.TrimSpace(req.User) != "" {
		return nil, http.StatusBadRequest, "a thread keeps its user"
	}
	caller := access.UserFrom(r.Context())
	thread, err := h.opts.Board.Notes.Get(r.Context(), req.Parent)
	if err != nil {
		return nil, statusOf(err), err.Error()
	}
	if thread.UserID == caller.ID {
		return caller, 0, ""
	}
	if !caller.CanAdmin() {
		return nil, http.StatusForbidden, "a thread keeps its user: " + thread.UserName + " or an admin replies"
	}
	if h.opts.Users == nil {
		return caller, 0, ""
	}
	user, err := h.opts.Users.Get(r.Context(), thread.UserID)
	if err != nil {
		return caller, 0, ""
	}
	return user, 0, ""
}
