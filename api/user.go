package api

import (
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
)

// UserRequest creates or changes a user. An empty password on an update
// keeps the one there is.
type UserRequest struct {
	Name     string   `json:"name"`
	Password string   `json:"password"`
	Role     string   `json:"role"`
	Folders  []string `json:"folders"`
}

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

// createUser adds a user with a hashed password.
func (h *Handlers) createUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var req UserRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.Password == "" {
		fail(w, http.StatusBadRequest, "a user needs a name and a password")
		return
	}
	hash, err := access.HashPassword(req.Password)
	if err != nil {
		failErr(w, err)
		return
	}
	user := &model.User{Name: req.Name, Password: hash, Role: req.Role, Folders: model.EncodeList(req.Folders)}
	if user.Role == "" {
		user.Role = model.RoleViewer
	}
	if err := h.opts.Users.Create(r.Context(), user); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

// updateUser changes a user's role, folders, or password.
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
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Folders != nil {
		user.Folders = model.EncodeList(req.Folders)
	}
	if req.Password != "" {
		hash, err := access.HashPassword(req.Password)
		if err != nil {
			failErr(w, err)
			return
		}
		user.Password = hash
	}
	if err := h.opts.Users.Update(r.Context(), user); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// deleteUser removes a user. Not yourself: the instance keeps an admin.
func (h *Handlers) deleteUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	id := param(r, "id")
	if id == access.UserFrom(r.Context()).ID {
		fail(w, http.StatusBadRequest, "you cannot delete yourself")
		return
	}
	if _, err := h.opts.Users.Get(r.Context(), id); err != nil {
		failErr(w, err)
		return
	}
	if err := h.opts.Users.Delete(r.Context(), id); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
