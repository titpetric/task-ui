package api

import (
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
)

// runAsUser returns the user a workflow's runs run as: the one its
// run-as names, else its owner; storage.ErrNotFound for a user nobody
// holds any more.
func (h *Handlers) runAsUser(r *http.Request, wf *model.Workflow) (*model.User, error) {
	if h.opts.Users == nil {
		return &model.User{ID: wf.RunsAs(), Name: wf.UserName}, nil
	}
	return h.opts.Users.Get(r.Context(), wf.RunsAs())
}

// runAsName names the user a workflow's runs run as, "" for a user
// nobody holds any more.
func (h *Handlers) runAsName(r *http.Request, wf *model.Workflow) string {
	if wf.RunAsUserID == "" {
		return wf.UserName
	}
	user, err := h.runAsUser(r, wf)
	if err != nil {
		return ""
	}
	return user.Name
}

// pickRunAs resolves the user a workflow request names to run as, by
// name. Nil for no name, and the caller for their own; only an admin
// names someone else (403), and a name nobody holds is a 404.
func (h *Handlers) pickRunAs(r *http.Request, name string) (*model.User, int, string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, 0, ""
	}
	caller := access.UserFrom(r.Context())
	if name == caller.Name {
		return caller, 0, ""
	}
	if !caller.CanAdmin() {
		return nil, http.StatusForbidden, "an admin picks another user for a workflow to run as"
	}
	if h.opts.Users == nil {
		return nil, http.StatusNotFound, "no user " + name
	}
	user, err := h.opts.Users.GetByName(r.Context(), name)
	if err != nil {
		return nil, http.StatusNotFound, "no user " + name
	}
	return user, 0, ""
}

// mayRunAs returns the first step the user a workflow runs as may not
// run, "" when they may run them all, as mayRun does for the caller.
func (h *Handlers) mayRunAs(r *http.Request, user *model.User, steps []model.WorkflowStep) string {
	if h.opts.Users == nil {
		return h.mayRun(r, steps)
	}
	return h.opts.Flows.MayRun(r.Context(), h.opts.Users.Permissions(r.Context(), user), steps)
}
