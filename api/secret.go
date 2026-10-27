package api

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// SecretRequest sets a variable: a secret, whose value no page shows, or
// an env entry, whose value it does. An empty is_secret makes a secret;
// an ID replaces that row, so its scope and name may change too.
type SecretRequest struct {
	ID       string `json:"id"`
	Scope    string `json:"scope"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	IsSecret *bool  `json:"is_secret"`
}

// SecretView is a variable as the API shows it: a secret never shows its
// content, an env entry does. The list adds how many runs read it, in the
// log's newest storage.SecretReadKeep, and the last of them.
type SecretView struct {
	ID       string            `json:"id"`
	Scope    string            `json:"scope"`
	Name     string            `json:"name"`
	IsSecret bool              `json:"is_secret"`
	Value    string            `json:"value,omitempty"`
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
		view := SecretView{ID: secret.ID, Scope: secret.Scope, Name: secret.Name, IsSecret: secret.IsSecret != 0, Reads: counts[secret.ID]}
		if !view.IsSecret {
			view.Value = secret.Content
		}
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

// setSecret creates a variable or replaces its content. Without a
// scope it reaches every workspace; an ID replaces that row.
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
	if req.Scope == "" {
		req.Scope = storage.ScopeAll
	}
	if err := h.secretScope(r, req.Scope); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	isSecret := req.IsSecret == nil || *req.IsSecret
	secret := &model.Secret{ID: req.ID, Scope: req.Scope, Name: strings.TrimSpace(req.Name), Content: req.Content}
	if isSecret {
		secret.IsSecret = 1
	}
	if req.ID == "" {
		set := h.opts.Secrets.Set
		if !isSecret {
			set = h.opts.Secrets.SetEnv
		}
		if err := set(r.Context(), secret); err != nil {
			failErr(w, err)
			return
		}
	} else {
		existing, err := h.opts.Secrets.ByID(r.Context(), req.ID)
		if err != nil {
			failErr(w, err)
			return
		}
		// An empty value keeps a secret's, whatever the box now says,
		// since an edit of one starts blank; an env entry may be empty.
		if existing.IsSecret != 0 && req.Content == "" {
			secret.Content = existing.Content
		}
		secret.CreatedAt = existing.CreatedAt
		if err := h.opts.Secrets.Update(r.Context(), secret); err != nil {
			failErr(w, err)
			return
		}
	}
	view := SecretView{ID: secret.ID, Scope: secret.Scope, Name: secret.Name, IsSecret: isSecret}
	if !isSecret {
		view.Value = secret.Content
	}
	writeJSON(w, http.StatusOK, view)
}

// secretScope refuses a scope that is not every workspace, all,
// workspace:<folder>, or a folder of older rows.
func (h *Handlers) secretScope(r *http.Request, scope string) error {
	if scope == storage.ScopeAll {
		return nil
	}
	if name, ok := storage.ScopeWorkspaceName(scope); ok {
		if h.opts.Workspaces == nil {
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

// rekeySecrets writes every secret again for the instance's current
// encryption key.
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
