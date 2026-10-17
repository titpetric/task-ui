package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// secrets lists the secrets, values masked.
func (h *Handlers) secrets(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	p := h.page(r, "Secrets")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Secrets")
	p.Secrets, _ = h.opts.Secrets.List(r.Context())
	h.render(w, http.StatusOK, "secrets", p)
}

// secretNew shows the form for a new secret, scoped by the query.
func (h *Handlers) secretNew(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = storage.ScopeAll
	}
	p := h.page(r, "Add secret")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Secrets", "/settings/secrets", "Add")
	p.Secret = &model.Secret{Scope: scope}
	h.render(w, http.StatusOK, "secret_form", p)
}

// secretEdit shows the form to replace a secret's value.
func (h *Handlers) secretEdit(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	q := r.URL.Query()
	secret, err := h.opts.Secrets.Get(r.Context(), model.Secret{Scope: q.Get("scope"), Name: q.Get("name")})
	if err != nil {
		h.missing(w, r, "No secret named "+q.Get("name")+" for "+q.Get("scope")+".")
		return
	}
	p := h.page(r, "Replace "+secret.Name)
	p.Crumbs = view.Crumbs("Settings", "/settings", "Secrets", "/settings/secrets", secret.Name)
	p.Secret = &model.Secret{ID: secret.ID, Scope: secret.Scope, Name: secret.Name}
	h.render(w, http.StatusOK, "secret_form", p)
}
