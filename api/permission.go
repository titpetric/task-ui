package api

import (
	"errors"
	"net/http"
	"sort"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// getPermissions answers permissions.yml, for the permission matrix.
// Takes user/view.
func (h *Handlers) getPermissions(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "user/view") {
		return
	}
	p := config.Default()
	writeJSON(w, http.StatusOK, PermissionsView{Rules: p.Rules, Sections: p.Sections, Permissions: p.Permissions, Paths: p.Paths(), Roles: model.Roles()})
}

// getRolePolicy answers a role's policy. Takes user/view.
func (h *Handlers) getRolePolicy(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "user/view") {
		return
	}
	role := param(r, "role")
	if !model.ValidRole(role) {
		fail(w, http.StatusNotFound, "no role "+role)
		return
	}
	h.writeRolePolicy(w, r, role)
}

// putRolePolicy sets entries of a role's policy, {path: "allow" | "deny"
// | ""}, "" clearing one. Takes user/edit, and for your own role user/edit
// through it; the admin role holds no policy (409).
func (h *Handlers) putRolePolicy(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "user/edit") {
		return
	}
	role := param(r, "role")
	if !model.ValidRole(role) {
		fail(w, http.StatusNotFound, "no role "+role)
		return
	}
	if role == model.RoleAdmin {
		fail(w, http.StatusConflict, "the admin role allows everything and holds no policy")
		return
	}
	if role == access.UserFrom(r.Context()).Role && !h.editsUsersByRole(r) {
		fail(w, http.StatusForbidden, "changing your own role's policy takes user/edit through that role")
		return
	}
	if !h.putPolicy(w, r, permission.KindRole, role) {
		return
	}
	h.writeRolePolicy(w, r, role)
}

// getUserPolicy answers a user's policy, by name, with the role's.
// Takes user/view.
func (h *Handlers) getUserPolicy(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "user/view") {
		return
	}
	user, err := h.opts.Users.GetByName(r.Context(), param(r, "name"))
	if err != nil {
		failErr(w, err)
		return
	}
	h.writeUserPolicy(w, r, user)
}

// putUserPolicy sets entries of a user's policy, as putRolePolicy does a
// role's. Takes user/edit; your own takes user/edit through your role,
// so a user policy never grants itself more. An admin's is not
// editable (409).
func (h *Handlers) putUserPolicy(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "user/edit") {
		return
	}
	user, err := h.opts.Users.GetByName(r.Context(), param(r, "name"))
	if err != nil {
		failErr(w, err)
		return
	}
	if user.CanAdmin() {
		fail(w, http.StatusConflict, "an admin may do everything and holds no policy")
		return
	}
	if user.Name == access.UserFrom(r.Context()).Name && !h.editsUsersByRole(r) {
		fail(w, http.StatusForbidden, "changing your own policy takes user/edit through your role")
		return
	}
	if !h.putPolicy(w, r, permission.KindUser, user.Name) {
		return
	}
	h.writeUserPolicy(w, r, user)
}

// editsUsersByRole reports whether the user of a request holds user/edit
// through the role, not only through a policy of the user's own: an
// admin, or a role that allows it.
func (h *Handlers) editsUsersByRole(r *http.Request) bool {
	set := h.permissions(r)
	return set.User.CanAdmin() || (set.Allowed("user/edit") && set.Role.Effect("user/edit") == permission.Allow)
}

// errEffect reports an effect a policy entry cannot take.
var errEffect = errors.New(`an effect is "allow", "deny" or "" to inherit`)

// putPolicy reads {path: effect} and writes each entry of a subject,
// after checking them all, so a bad entry changes nothing. It answers
// 400 for a path or an effect it does not know, and reports success.
func (h *Handlers) putPolicy(w http.ResponseWriter, r *http.Request, kind, subject string) bool {
	var req map[string]string
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return false
	}
	paths := make([]string, 0, len(req))
	for path, effect := range req {
		if !config.Default().Valid(path) {
			fail(w, http.StatusBadRequest, config.ErrUnknown.Error()+": "+path)
			return false
		}
		if effect != "" && effect != permission.Allow && effect != permission.Deny {
			fail(w, http.StatusBadRequest, errEffect.Error())
			return false
		}
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := h.opts.Users.Policies().Put(r.Context(), kind, subject, path, req[path]); err != nil {
			fail(w, http.StatusBadRequest, err.Error())
			return false
		}
	}
	return true
}

// writeRolePolicy answers a role's policy and what the role may do.
func (h *Handlers) writeRolePolicy(w http.ResponseWriter, r *http.Request, role string) {
	view := PolicyView{Kind: permission.KindRole, Subject: role, Fixed: role == model.RoleAdmin, Policy: permission.Policy{}}
	if !view.Fixed {
		policy, err := h.opts.Users.Policies().Policy(r.Context(), permission.KindRole, role)
		if err != nil {
			failErr(w, err)
			return
		}
		view.Policy = policy
	}
	view.Effective = permission.Set{User: &model.User{Role: role}, Role: view.Policy}.Effective()
	writeJSON(w, http.StatusOK, view)
}

// writeUserPolicy answers a user's policy, the role's, and what the user
// may do by both.
func (h *Handlers) writeUserPolicy(w http.ResponseWriter, r *http.Request, user *model.User) {
	view := PolicyView{Kind: permission.KindUser, Subject: user.Name, Role: user.Role, Fixed: user.CanAdmin(), Policy: permission.Policy{}}
	set := h.opts.Users.Permissions(r.Context(), user)
	if !view.Fixed && set.Own != nil {
		view.Policy, view.RolePolicy = set.Own, set.Role
	}
	view.Effective = set.Effective()
	writeJSON(w, http.StatusOK, view)
}
