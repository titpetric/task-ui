package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/integration"
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
		if integration.Internal(secret.Scope) {
			continue
		}
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
	if req.Scope == "" || req.Scope == storage.ScopeAll {
		req.Scope = storage.SharedScope(model.Default)
	}
	if err := h.secretScope(r, req.Scope); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if problem := h.secretKey(r, req.Scope); problem != "" {
		fail(w, http.StatusConflict, problem)
		return
	}
	secret := &model.Secret{Scope: req.Scope, Name: strings.TrimSpace(req.Name), Content: req.Content}
	if err := h.opts.Secrets.Set(r.Context(), secret); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, SecretView{ID: secret.ID, Scope: secret.Scope, Name: secret.Name})
}

// secretScope refuses a scope that is not a workspace's shared scope,
// workspace:<folder>, or a folder.
func (h *Handlers) secretScope(r *http.Request, scope string) error {
	if name, ok := strings.CutPrefix(scope, storage.ScopeWorkspace); ok {
		if name == model.Default || h.opts.Workspaces == nil {
			return nil
		}
		if _, err := h.opts.Workspaces.GetByFolder(r.Context(), name); err != nil {
			return errors.New("unknown workspace " + name)
		}
		return nil
	}
	if h.opts.Catalogue.Find(scope) == nil {
		return errors.New("unknown folder " + scope)
	}
	return nil
}

// secretKey says why a scope's workspace stores no secret: it has no
// encryption key yet. "" when it has one, or keeps no keys of its own.
func (h *Handlers) secretKey(r *http.Request, scope string) string {
	if h.opts.WorkspaceKeys == nil || h.opts.Workspaces == nil || h.opts.Secrets == nil || h.opts.Secrets.Workspaces == nil {
		return ""
	}
	name, err := h.opts.Secrets.Workspace(r.Context(), scope)
	if err != nil {
		return ""
	}
	row, err := h.opts.Workspaces.GetByFolder(r.Context(), name)
	if err != nil || h.opts.WorkspaceKeys.EncryptKey(row.ID) != nil {
		return ""
	}
	return "the workspace " + name + " has no encryption key, so it stores no secrets: generate or upload one in Settings, Keys"
}

// rekeySecrets writes every secret again for the current key of its
// scope: a workspace's, or the instance key for the services'.
func (h *Handlers) rekeySecrets(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	n, err := h.opts.Secrets.Rekey(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rekeyed": n, "keys": h.opts.Secrets.Vault.Recipients()})
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
