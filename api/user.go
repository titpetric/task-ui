package api

import (
	"net/http"

	"github.com/titpetric/task-ui/model"
)

// UserRequest changes what the instance keeps about a user: the folders
// they see. Names, passwords and roles come from the password file and
// the role lists, not from here.
type UserRequest struct {
	Folders []string `json:"folders"`
}

// managedElsewhere is the reply to making or removing users here.
const managedElsewhere = "users are managed in the password file: task-ui passwd <user> adds or resets one, removing the line removes the user; TASK_UI_ADMINS and TASK_UI_VIEWERS set the roles"

// listUsers returns every user. Admins only.
func (h *Handlers) listUsers(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	users, err := h.opts.Users.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

// createUser says where to make users.
func (h *Handlers) createUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	fail(w, http.StatusMethodNotAllowed, managedElsewhere)
}

// updateUser changes the folders a user sees. Empty shows them all.
func (h *Handlers) updateUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	user, err := h.opts.Users.Get(r.Context(), param(r, "id"))
	if err != nil {
		failErr(w, err)
		return
	}
	var req UserRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	user.Folders = model.EncodeList(req.Folders)
	if err := h.opts.Users.Update(r.Context(), user); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// deleteUser says where to remove users.
func (h *Handlers) deleteUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	fail(w, http.StatusMethodNotAllowed, managedElsewhere)
}
