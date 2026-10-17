package api

import (
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// SecretRequest sets a secret.
type SecretRequest struct {
	Scope   string `json:"scope"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// SecretView is a secret as the API shows it: never its content.
type SecretView struct {
	ID    string `json:"id"`
	Scope string `json:"scope"`
	Name  string `json:"name"`
}

// listSecrets returns every secret, content withheld. Admins only.
func (h *Handlers) listSecrets(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	secrets, err := h.opts.Secrets.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	views := []SecretView{}
	for _, secret := range secrets {
		views = append(views, SecretView{ID: secret.ID, Scope: secret.Scope, Name: secret.Name})
	}
	writeJSON(w, http.StatusOK, map[string]any{"secrets": views})
}

// setSecret creates a secret or replaces its content.
func (h *Handlers) setSecret(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var req SecretRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Scope = strings.TrimSpace(req.Scope)
	if req.Scope == "" {
		req.Scope = storage.ScopeAll
	}
	if req.Scope != storage.ScopeAll && h.opts.Catalogue.Find(req.Scope) == nil {
		fail(w, http.StatusBadRequest, "unknown folder "+req.Scope)
		return
	}
	secret := &model.Secret{Scope: req.Scope, Name: strings.TrimSpace(req.Name), Content: req.Content}
	if err := h.opts.Secrets.Set(r.Context(), secret); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, SecretView{ID: secret.ID, Scope: secret.Scope, Name: secret.Name})
}

// deleteSecret removes a secret.
func (h *Handlers) deleteSecret(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	if err := h.opts.Secrets.Delete(r.Context(), param(r, "id")); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
