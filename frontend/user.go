package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/frontend/view"
)

// users lists who may sign in.
func (h *Handlers) users(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	p := h.page(r, "Users")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Users")
	p.Users, _ = h.opts.Users.List(r.Context())
	h.render(w, http.StatusOK, "users", p)
}

// userEdit shows the folders of a user, to change.
func (h *Handlers) userEdit(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
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
	h.render(w, http.StatusOK, "user_form", p)
}
