package frontend

import (
	"context"
	"net/http"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// secrets lists the secrets, values masked.
func (h *Handlers) secrets(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	p := h.page(r, "Secrets")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Secrets")
	p.Encryption = h.opts.Secrets.Vault.Recipients()
	secrets, err := h.opts.Secrets.List(r.Context())
	p.Secrets = userSecrets(secrets)
	if err != nil {
		p.Error = err.Error()
	}
	p.SecretGroups = h.secretGroups(r.Context(), p.Secrets, p.Folders)
	h.render(w, http.StatusOK, "secrets", p)
}

// secretNew shows the form for a new secret, scoped by the query.
func (h *Handlers) secretNew(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	scope := r.URL.Query().Get("scope")
	if scope == "" || scope == storage.ScopeAll {
		scope = storage.SharedScope(model.Default)
	}
	p := h.page(r, "Add secret")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Secrets", "/settings/secrets", "Add")
	p.Encryption = h.opts.Secrets.Vault.Recipients()
	p.Secret = &model.Secret{Scope: scope}
	p.SecretGroups = h.secretGroups(r.Context(), nil, p.Folders)
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
	p.Encryption = h.opts.Secrets.Vault.Recipients()
	p.Crumbs = view.Crumbs("Settings", "/settings", "Secrets", "/settings/secrets", secret.Name)
	p.Secret = &model.Secret{ID: secret.ID, Scope: secret.Scope, Name: secret.Name}
	p.SecretGroups = h.secretGroups(r.Context(), nil, p.Folders)
	h.render(w, http.StatusOK, "secret_form", p)
}

// secretGroups sorts the secrets and the folders by workspace, in the
// order of the workspaces, the default one first.
func (h *Handlers) secretGroups(ctx context.Context, secrets []model.Secret, folders []*folder.Folder) []view.SecretGroup {
	var groups []view.SecretGroup
	index := map[string]int{}
	add := func(name, title string, key *sshkey.Key) int {
		if i, ok := index[name]; ok {
			return i
		}
		if title == "" {
			title = name
		}
		index[name] = len(groups)
		groups = append(groups, view.SecretGroup{Workspace: name, Title: title, Key: key})
		return index[name]
	}
	add(model.Default, "Default", nil)
	if h.opts.Workspaces != nil {
		workspaces, _ := h.opts.Workspaces.List(ctx)
		for _, ws := range workspaces {
			var key *sshkey.Key
			if h.opts.WorkspaceKeys != nil {
				key = h.opts.WorkspaceKeys.EncryptKey(ws.ID)
			}
			i := add(ws.Folder, ws.Title, key)
			groups[i].Key = key
		}
	}
	for _, secret := range secrets {
		name, _ := h.opts.Secrets.Workspace(ctx, secret.Scope)
		i := add(name, "", nil)
		groups[i].Secrets = append(groups[i].Secrets, secret)
	}
	for _, f := range folders {
		name, _ := h.opts.Secrets.Workspace(ctx, f.ID)
		i := add(name, "", nil)
		groups[i].Folders = append(groups[i].Folders, f.ID)
	}
	return groups
}

// userSecrets leaves out the secrets of the integrations, which their own
// page shows.
func userSecrets(all []model.Secret) []model.Secret {
	secrets := []model.Secret{}
	for _, secret := range all {
		if !integration.Internal(secret.Scope) {
			secrets = append(secrets, secret)
		}
	}
	return secrets
}
