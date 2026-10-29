package api

import (
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// pickRunAs resolves the user a request names to run as, by name, for a
// workflow or a prompt. Nil for no name, and the caller for their own;
// only an admin names someone else (403), and a name nobody holds is a
// 404.
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
		return nil, http.StatusForbidden, "an admin picks another user to run as"
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

// permissionsOf returns what a user may do, the caller's permissions without a user store.
func (h *Handlers) permissionsOf(r *http.Request, user *model.User) permission.Set {
	if h.opts.Users == nil {
		return h.permissions(r)
	}
	return h.opts.Users.Permissions(r.Context(), user)
}
