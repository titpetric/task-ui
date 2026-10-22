package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
)

// users lists who may sign in.
func (h *Handlers) users(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "user/view") {
		return
	}
	p := h.page(r, "Users")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Users")
	p.Users, _ = h.opts.Users.List(r.Context())
	p.CanAddUsers = h.opts.AddUsers
	h.render(w, http.StatusOK, "users", p)
}

// userNew shows the form that makes a user in the password file.
func (h *Handlers) userNew(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "user/edit") {
		return
	}
	if !h.opts.AddUsers {
		h.missing(w, r, "task-ui has no password file to add users to.")
		return
	}
	p := h.page(r, "Create user")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Users", "/settings/users", "Create")
	p.FolderChoices = view.FolderChoices(h.opts.Catalogue.List(), nil)
	h.render(w, http.StatusOK, "user_form", p)
}

// userEdit shows the folders of a user, to change, and their MCP
// access. Your own page opens without user/view, for your MCP access.
func (h *Handlers) userEdit(w http.ResponseWriter, r *http.Request) {
	me := access.UserFrom(r.Context())
	own := me != nil && me.ID == param(r, "id")
	if !own && !h.require(w, r, "user/view") {
		return
	}
	user, err := h.opts.Users.Get(r.Context(), param(r, "id"))
	if err != nil {
		h.missing(w, r, "No user with that ID.")
		return
	}
	p := h.page(r, "Folders of "+user.Name)
	p.Crumbs = view.Crumbs("Settings", "/settings", "Users", "/settings/users", user.Name)
	p.Account = user
	p.FolderChoices = view.FolderChoices(h.opts.Catalogue.List(), user.FolderList())
	p.AccountMCP = h.mcpAccess(r, user, own)
	h.render(w, http.StatusOK, "user_form", p)
}

// mcpAccess returns a user's MCP token as their page shows it.
func (h *Handlers) mcpAccess(r *http.Request, user *model.User, own bool) view.MCPAccess {
	if h.opts.Guard == nil || h.opts.Guard.Tokens == nil {
		return view.MCPAccess{}
	}
	token, err := h.opts.Guard.Tokens.Get(r.Context(), user.ID, access.MCPToken)
	if err != nil {
		token = nil
	}
	a := view.NewMCPAccess(token)
	a.CanEnable = h.can(r, "user/edit")
	a.CanManage = a.CanEnable || own && a.Enabled
	return a
}
