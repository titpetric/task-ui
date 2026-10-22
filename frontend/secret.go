package frontend

import (
	"context"
	"net/http"
	"strconv"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// secrets lists the secrets by workspace, values withheld; ?workspace=
// shows one workspace alone.
func (h *Handlers) secrets(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "secret/view") {
		return
	}
	p := h.page(r, "Secrets")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Secrets")
	secrets, err := h.opts.Secrets.List(r.Context())
	p.Secrets = userSecrets(secrets)
	if err != nil {
		p.Error = err.Error()
	}
	p.SecretGroups = h.secretGroups(r.Context(), p.Secrets, p.Folders)
	if only := r.URL.Query().Get("workspace"); only != "" {
		for _, group := range p.SecretGroups {
			if group.Workspace == only {
				p.SecretWorkspace, p.SecretGroups = only, []view.SecretGroup{group}
				break
			}
		}
	}
	h.render(w, http.StatusOK, "secrets", p)
}

// rekeyedNotice is the notice after a rotation wrote n secrets again,
// or n of what, such as integration settings; "" without a count.
func rekeyedNotice(n, what string) string {
	count, err := strconv.Atoi(n)
	if err != nil || count < 0 {
		return ""
	}
	if what != "integration setting" {
		what = "secret"
	}
	if count != 1 {
		what += "s"
	}
	return "Encryption key rotated. Re-encrypted " + strconv.Itoa(count) + " " + what + "."
}

// secretNew shows the form for a new secret, scoped by the query.
func (h *Handlers) secretNew(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "secret/edit") {
		return
	}
	scope := r.URL.Query().Get("scope")
	if scope == "" || scope == storage.ScopeAll {
		scope = storage.SharedScope(model.Default)
	}
	p := h.page(r, "Add secret")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Secrets", "/settings/secrets", "Add")
	p.SecretGroups = h.secretGroups(r.Context(), nil, p.Folders)
	p.Secret = &model.Secret{Scope: keyedScope(p.SecretGroups, scope, h.workspaceOf(r.Context(), scope))}
	h.render(w, http.StatusOK, "secret_form", p)
}

// workspaceOf returns the folder of the workspace a scope is in.
func (h *Handlers) workspaceOf(ctx context.Context, scope string) string {
	name, _ := h.opts.Secrets.Workspace(ctx, scope)
	return name
}

// keyedScope keeps a scope whose workspace has an encryption key, else
// picks the shared scope of the first workspace with one: a workspace
// without a key stores no secret.
func keyedScope(groups []view.SecretGroup, scope, workspace string) string {
	for _, group := range groups {
		if group.Workspace == workspace && group.Key != nil {
			return scope
		}
	}
	for _, group := range groups {
		if group.Key != nil {
			return group.Shared()
		}
	}
	return scope
}

// secretEdit shows the form to replace a secret's value.
func (h *Handlers) secretEdit(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "secret/view") {
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
