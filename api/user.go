package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// UserRequest makes a user, with a name and a password the password file
// takes, or changes the folders a user sees. Roles come from the role
// lists.
type UserRequest struct {
	// Name and Password make a new user; an edit takes only the folders.
	Name     string   `json:"name,omitempty"`
	Password string   `json:"password,omitempty"`
	Folders  []string `json:"folders"`
}

// managedElsewhere is the reply to making or removing users here.
const managedElsewhere = "users are managed in the password file: task-ui passwd <user> adds or resets one, removing the line removes the user; TASK_UI_ADMINS and TASK_UI_VIEWERS set the roles"

// listUsers returns every user. Takes user/view.
func (h *Handlers) listUsers(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "user/view") {
		return
	}
	users, err := h.opts.Users.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users})
}

// ErrUserExists is what AddUser reports for a name the password file
// already holds.
var ErrUserExists = errors.New("the password file already has that user")

// createUser adds a user to the password file, with their folders, when
// the server has one to write. Takes user/edit.
func (h *Handlers) createUser(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "user/edit") {
		return
	}
	if h.opts.AddUser == nil {
		fail(w, http.StatusMethodNotAllowed, managedElsewhere)
		return
	}
	var req UserRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	name := strings.TrimSpace(req.Name)
	if err := h.opts.AddUser(r.Context(), name, req.Password); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrUserExists) {
			status = http.StatusConflict
		}
		fail(w, status, err.Error())
		return
	}
	user, err := h.opts.Users.GetByName(r.Context(), name)
	if err != nil {
		failErr(w, err)
		return
	}
	if len(req.Folders) > 0 {
		user.Folders = model.EncodeList(req.Folders)
		if err := h.opts.Users.Update(r.Context(), user); err != nil {
			failErr(w, err)
			return
		}
	}
	writeJSON(w, http.StatusCreated, user)
}

// updateUser changes the folders a user sees. Empty shows them all.
func (h *Handlers) updateUser(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "user/edit") {
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
	if !h.require(w, r, "user/delete") {
		return
	}
	fail(w, http.StatusMethodNotAllowed, managedElsewhere)
}
