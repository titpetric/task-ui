package frontend

import (
	"context"
	"net/http"
	"strconv"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
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
	p.Step = view.StepSecrets
	secrets, err := h.opts.Secrets.List(r.Context())
	p.Secrets = userSecrets(secrets)
	if err != nil {
		p.Error = err.Error()
	}
	p.SecretGroups = h.secretGroups(r.Context(), p.Secrets, p.Folders)
	if h.opts.SecretReads != nil {
		p.SecretReadCounts, _ = h.opts.SecretReads.Counts(r.Context())
		p.SecretLastReads, _ = h.opts.SecretReads.Last(r.Context())
	}
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

// workspaceSecrets lists the secrets of one workspace, values withheld,
// with Add secret and Replace the value in dialogs on the page. A
// workspace unknown is not found.
func (h *Handlers) workspaceSecrets(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "secret/view") {
		return
	}
	ctx, name := r.Context(), param(r, "folder")
	var workspace *model.Workspace
	if h.opts.Workspaces != nil {
		workspace, _ = h.opts.Workspaces.GetByFolder(ctx, name)
	}
	if workspace == nil {
		h.missing(w, r, "No workspace named "+name+".")
		return
	}
	p := h.page(r, "Secrets of "+workspace.Folder)
	p.Crumbs = view.Crumbs("Settings", "/settings", "Workspaces", "/settings/workspaces", workspace.Folder, "/settings/workspaces/"+workspace.Folder, "Secrets")
	p.Step = view.StepSecrets
	p.Workspace = workspace
	secrets, err := h.opts.Secrets.List(ctx)
	if err != nil {
		p.Error = err.Error()
	}
	for _, group := range h.secretGroups(ctx, userSecrets(secrets), p.Folders) {
		if group.Workspace == workspace.Folder {
			p.SecretGroups = []view.SecretGroup{group}
		}
	}
	p.Secret = &model.Secret{Scope: storage.SharedScope(workspace.Folder)}
	if h.opts.SecretReads != nil {
		p.SecretReadCounts, _ = h.opts.SecretReads.Counts(ctx)
		p.SecretLastReads, _ = h.opts.SecretReads.Last(ctx)
	}
	h.render(w, http.StatusOK, "workspace_secrets", p)
}

// secretReads lists the runs that got a secret, newest first, a page
// at a time; the value never shows.
func (h *Handlers) secretReads(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "secret/view") {
		return
	}
	secrets, err := h.opts.Secrets.Secrets.List(r.Context())
	if err != nil {
		h.missing(w, r, "No secret with that ID.")
		return
	}
	id := param(r, "id")
	var found *model.Secret
	for _, secret := range userSecrets(secrets) {
		if secret.ID == id {
			found = &model.Secret{ID: secret.ID, Scope: secret.Scope, Name: secret.Name}
		}
	}
	if found == nil {
		h.missing(w, r, "No secret with that ID.")
		return
	}
	p := h.page(r, "Reads of "+found.Name)
	p.Crumbs = view.Crumbs("Settings", "/settings", "Secrets", "/settings/secrets", found.Name)
	p.Step = view.StepSecrets
	p.Secret = found
	if h.opts.SecretReads != nil {
		q := r.URL.Query()
		before, after := q.Get("before"), q.Get("after")
		reads, err := h.opts.SecretReads.List(r.Context(), id, before, after, secretReadPage+1)
		if err != nil {
			p.Error = err.Error()
		}
		var prev, next string
		p.SecretReads, prev, next = paged(reads, secretReadPage, before, after, func(read model.SecretRead) string { return read.ID })
		p.Pager = view.NewPager("/settings/secrets/"+id+"/reads", q, prev, next)
	}
	h.render(w, http.StatusOK, "secret_reads", p)
}

// secretReadPage is how many reads a page of a secret's log shows.
const secretReadPage = 50

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
	p.Step = view.StepSecrets
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
	p.Step = view.StepSecrets
	p.Secret = &model.Secret{ID: secret.ID, Scope: secret.Scope, Name: secret.Name}
	p.SecretGroups = h.secretGroups(r.Context(), nil, p.Folders)
	h.render(w, http.StatusOK, "secret_form", p)
}

// secretGroups sorts the secrets and the folders by workspace, in the
// order of the workspaces, the default one first; each takes the
// instance's encryption key, which encrypts every secret.
func (h *Handlers) secretGroups(ctx context.Context, secrets []model.Secret, folders []*folder.Folder) []view.SecretGroup {
	var groups []view.SecretGroup
	var key *keystore.Key
	if h.opts.Keys != nil {
		key = h.opts.Keys.EncryptKey(ctx, storage.ScopeInstance)
	}
	index := map[string]int{}
	add := func(name, title string, key *keystore.Key) int {
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
	add(model.Default, "Default", key)
	if h.opts.Workspaces != nil {
		workspaces, _ := h.opts.Workspaces.List(ctx)
		for _, ws := range workspaces {
			add(ws.Folder, ws.Title, key)
		}
	}
	for _, secret := range secrets {
		name, _ := h.opts.Secrets.Workspace(ctx, secret.Scope)
		i := add(name, "", key)
		groups[i].Secrets = append(groups[i].Secrets, secret)
	}
	for _, f := range folders {
		name, _ := h.opts.Secrets.Workspace(ctx, f.ID)
		i := add(name, "", key)
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
