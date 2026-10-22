package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/frontend/view"
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

// userEdit shows the folders of a user, to change.
func (h *Handlers) userEdit(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "user/view") {
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
	h.render(w, http.StatusOK, "user_form", p)
}
