package frontend

import (
	"net/http"
	"net/url"

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
	p.Step = view.StepUsers
	p.Users, _ = h.opts.Users.List(r.Context())
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
	p := h.page(r, "Add user")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Users", "/settings/users", "Add user")
	p.Step = view.StepUsers
	p.FolderChoices = view.FolderChoices(h.opts.Catalogue.List(), nil)
	h.render(w, http.StatusOK, "user_form", p)
}

// userEdit holds the forms of a user, their roles, folders, permissions, keys and MCP access, which change through the API with user/edit.
// It opens as the profile does, for your own or with user/view.
func (h *Handlers) userEdit(w http.ResponseWriter, r *http.Request) {
	user, own, ok := h.userPage(w, r)
	if !ok {
		return
	}
	p := h.page(r, "Edit "+user.Name)
	p.Crumbs = view.Crumbs(user.Name, profileHref(user.Name), "Edit")
	if p.Can("user/view") {
		p.Crumbs = view.Crumbs("Settings", "/settings", "Users", "/settings/users", user.Name, profileHref(user.Name), "Edit")
	}
	p.Account = user
	p.FolderChoices = view.FolderChoices(h.opts.Catalogue.List(), user.FolderList())
	p.AccountMCP = h.mcpAccess(r, user, own)
	if own || p.Can("key/view") {
		p.Keys = h.userKeys(r.Context(), user)
	}
	p.Profile = &view.Profile{Own: own, Matrix: h.userMatrix(r, p, user, own)}
	h.render(w, http.StatusOK, "user_edit", p)
}

// userByID sends the settings link of a user, by ID, to their edit page; your own opens without user/view.
// An ID no user holds answers not found.
func (h *Handlers) userByID(w http.ResponseWriter, r *http.Request) {
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
	http.Redirect(w, r, editHref(user.Name), http.StatusFound)
}

// profileHref returns the path of a user's profile.
func profileHref(name string) string {
	return "/users/" + url.PathEscape(name)
}

// editHref returns the path of a user's edit page.
func editHref(name string) string {
	return profileHref(name) + "/edit"
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
	a.Agent = user.Mcp
	a.CanEnable = h.can(r, "user/edit") && user.Mcp
	a.CanManage = a.CanEnable || own && a.Enabled
	return a
}
