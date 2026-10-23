package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// UserRequest makes a user, with a name and a password the password file
// takes, or changes the folders a user sees, and, by an admin, the admin
// flag, the role, MCP agent user and Run instance Claude when mentioned.
type UserRequest struct {
	// Name and Password make a new user; an edit takes the rest.
	Name     string   `json:"name,omitempty"`
	Password string   `json:"password,omitempty"`
	Folders  []string `json:"folders"`
	// Admin, Role, MCP and AgentReply change what the user is; absent
	// keeps the value. Role is runner or viewer.
	Admin      *bool  `json:"admin,omitempty"`
	Role       string `json:"role,omitempty"`
	MCP        *bool  `json:"mcp,omitempty"`
	AgentReply *bool  `json:"agent_reply,omitempty"`
}

// flags reports whether the request changes what the user is.
func (req *UserRequest) flags() bool {
	return req.Admin != nil || req.Role != "" || req.MCP != nil || req.AgentReply != nil
}

// managedElsewhere is the reply to making or removing users without a
// password file to write.
const managedElsewhere = "users are managed in the password file: task-ui passwd <user> adds or resets one, removing the line removes the user; an admin sets the roles on the user's profile or with task-ui user"

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
	views := make([]UserView, 0, len(users))
	for i := range users {
		views = append(views, h.userView(r.Context(), &users[i]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": views})
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
	writeJSON(w, http.StatusCreated, h.userView(r.Context(), user))
}

// updateUser changes the folders a user sees, empty shows them all, or
// what the user is: the admin flag, the role and the MCP flags, which
// only an admin changes. The last admin keeps the flag; MCP turned off
// ends the user's MCP token.
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
	if req.flags() {
		h.setFlags(w, r, user, &req)
		return
	}
	user.Folders = model.EncodeList(req.Folders)
	if err := h.opts.Users.Update(r.Context(), user); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.userView(r.Context(), user))
}

// deleteUser removes a user from the password file and storage, with
// their sessions, tokens and policy, when the server has a file to
// write. Takes user/delete; an admin and yourself stay.
func (h *Handlers) deleteUser(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "user/delete") {
		return
	}
	if h.opts.RemoveUser == nil {
		fail(w, http.StatusMethodNotAllowed, managedElsewhere)
		return
	}
	user, err := h.opts.Users.Get(r.Context(), param(r, "id"))
	if err != nil {
		failErr(w, err)
		return
	}
	switch {
	case user.ID == access.UserFrom(r.Context()).ID:
		fail(w, http.StatusForbidden, "you can't delete yourself")
		return
	case user.CanAdmin():
		fail(w, http.StatusForbidden, "you can't delete an admin here")
		return
	}
	if err := h.opts.RemoveUser(r.Context(), user.Name); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// setFlags changes what a user is, for an admin.
func (h *Handlers) setFlags(w http.ResponseWriter, r *http.Request, user *model.User, req *UserRequest) {
	if !access.UserFrom(r.Context()).CanAdmin() {
		fail(w, http.StatusForbidden, "only an admin changes the admin flag, the role and MCP")
		return
	}
	admin, role, mcp, reply := user.Admin, user.Role, user.Mcp, user.AgentReply
	if req.Admin != nil {
		admin = *req.Admin
	}
	if req.Role != "" {
		role = req.Role
	}
	if req.MCP != nil {
		mcp = *req.MCP
	}
	if req.AgentReply != nil {
		reply = *req.AgentReply
	}
	if !mcp && req.AgentReply == nil {
		reply = false
	}
	switch err := h.opts.Users.SetFlags(r.Context(), user, admin, role, mcp, reply, false); {
	case errors.Is(err, storage.ErrLastAdmin):
		fail(w, http.StatusConflict, "the last admin keeps the admin flag")
		return
	case err != nil:
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, h.userView(r.Context(), user))
}
