package api

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// SecretRequest sets a secret.
type SecretRequest struct {
	Scope   string `json:"scope"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// SecretView is a secret as the API shows it: never its content. The
// list adds how many runs read it, in the log's newest
// storage.SecretReadKeep, and the last of them.
type SecretView struct {
	ID       string            `json:"id"`
	Scope    string            `json:"scope"`
	Name     string            `json:"name"`
	Reads    int               `json:"reads"`
	LastRead *model.SecretRead `json:"last_read,omitempty"`
}

// listSecrets returns every secret, content withheld. Takes secret/view.
func (h *Handlers) listSecrets(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "secret/view") {
		return
	}
	secrets, err := h.opts.Secrets.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	counts, last := map[string]int{}, map[string]model.SecretRead{}
	if h.opts.SecretReads != nil {
		if counts, err = h.opts.SecretReads.Counts(r.Context()); err == nil {
			last, err = h.opts.SecretReads.Last(r.Context())
		}
		if err != nil {
			failErr(w, err)
			return
		}
	}
	views := []SecretView{}
	for _, secret := range secrets {
		if integration.Internal(secret.Scope) {
			continue
		}
		view := SecretView{ID: secret.ID, Scope: secret.Scope, Name: secret.Name, Reads: counts[secret.ID]}
		if read, ok := last[secret.ID]; ok {
			view.LastRead = &read
		}
		views = append(views, view)
	}
	writeJSON(w, http.StatusOK, map[string]any{"secrets": views})
}

// secretReads returns the log of a secret's reads, newest first: the
// runs that got it, ?limit= at a time (50 unless given), ?before= and
// ?after= a read's ID to page. Takes secret/view.
func (h *Handlers) secretReads(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "secret/view") {
		return
	}
	id := param(r, "id")
	secrets, err := h.opts.Secrets.Secrets.List(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	known := slices.ContainsFunc(secrets, func(secret model.Secret) bool {
		return secret.ID == id && !integration.Internal(secret.Scope)
	})
	if !known {
		fail(w, http.StatusNotFound, "no secret with that ID")
		return
	}
	reads := []model.SecretRead{}
	if h.opts.SecretReads != nil {
		q := r.URL.Query()
		limit := min(max(queryInt(r, "limit", 50), 1), 200)
		if reads, err = h.opts.SecretReads.List(r.Context(), id, q.Get("before"), q.Get("after"), limit); err != nil {
			failErr(w, err)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"reads": reads})
}

// setSecret creates a secret or replaces its content.
func (h *Handlers) setSecret(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "secret/edit") {
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
	if h.opts.Keys == nil || h.opts.Workspaces == nil || h.opts.Secrets == nil || h.opts.Secrets.Workspaces == nil {
		return ""
	}
	name, err := h.opts.Secrets.Workspace(r.Context(), scope)
	if err != nil {
		return ""
	}
	row, err := h.opts.Workspaces.GetByFolder(r.Context(), name)
	if err != nil || h.opts.Keys.EncryptKey(r.Context(), keystore.ScopeWorkspace(row.ID)) != nil {
		return ""
	}
	return "the workspace " + name + " has no encryption key, so it stores no secrets: generate or upload one in Settings, Keys"
}

// rekeySecrets writes every secret again for the current key of its
// scope: a workspace's, or the instance key for the services'.
func (h *Handlers) rekeySecrets(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "secret/edit") {
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
	if !h.require(w, r, "secret/delete") {
		return
	}
	if err := h.opts.Secrets.Delete(r.Context(), param(r, "id")); err != nil {
		failErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
