package frontend

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// secrets lists every variable runs may get, secrets and env entries
// together, values withheld for secrets; ?workspace= shows what reaches
// one workspace, its own and the ones of every workspace.
func (h *Handlers) secrets(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "secret/view") {
		return
	}
	ctx := r.Context()
	p := h.page(r, "Secrets")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Secrets")
	p.Step = view.StepSecrets
	secrets, err := h.opts.Secrets.List(ctx)
	p.Secrets = userSecrets(secrets)
	if err != nil {
		p.Error = err.Error()
	}
	p.SecretGroups = h.secretGroups(ctx)
	if h.opts.Keys != nil {
		p.SecretKey = h.opts.Keys.EncryptKey(ctx, storage.ScopeInstance)
	}
	p.SecretScopeTitles = scopeTitles(p.SecretGroups)
	if only := r.URL.Query().Get("workspace"); only != "" {
		p.SecretWorkspace = only
		p.Secrets = h.reachingWorkspace(ctx, p.Secrets, only)
	}
	if h.opts.SecretReads != nil {
		p.SecretReadCounts, _ = h.opts.SecretReads.Counts(ctx)
		p.SecretLastReads, _ = h.opts.SecretReads.Last(ctx)
	}
	p.IntegrationEnv = h.integrationEnv(r, p)
	h.render(w, http.StatusOK, "secrets", p)
}

// secretReads lists the runs that got a variable, newest first, a page
// at a time; a secret's value never shows.
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
			found = &model.Secret{ID: secret.ID, Scope: secret.Scope, Name: secret.Name, IsSecret: secret.IsSecret}
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

// secretNew shows the form for a new variable, scoped by the query.
func (h *Handlers) secretNew(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "secret/edit") {
		return
	}
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = storage.ScopeAll
	}
	p := h.page(r, "Add variable")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Secrets", "/settings/secrets", "Add variable")
	p.Step = view.StepSecrets
	p.SecretGroups = h.secretGroups(r.Context())
	p.Secret = &model.Secret{Scope: scope, IsSecret: 1}
	h.render(w, http.StatusOK, "secret_form", p)
}

// secretEdit shows the form to change a variable, by ID or by its scope
// and name.
func (h *Handlers) secretEdit(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "secret/view") {
		return
	}
	q := r.URL.Query()
	var secret *model.Secret
	var err error
	if id := q.Get("id"); id != "" {
		secret, err = h.opts.Secrets.ByID(r.Context(), id)
	} else {
		secret, err = h.opts.Secrets.Get(r.Context(), model.Secret{Scope: q.Get("scope"), Name: q.Get("name")})
	}
	if err != nil {
		h.missing(w, r, "No variable named "+q.Get("name")+" for "+q.Get("scope")+".")
		return
	}
	p := h.page(r, "Edit "+secret.Name)
	p.Crumbs = view.Crumbs("Settings", "/settings", "Secrets", "/settings/secrets", secret.Name)
	p.Step = view.StepSecrets
	p.Secret = secret
	p.SecretGroups = h.secretGroups(r.Context())
	h.render(w, http.StatusOK, "secret_form", p)
}

// secretGroups sorts the workspaces the form offers a scope for, the
// default one first while the instance has it, each with the instance's
// encryption key, which encrypts every variable.
func (h *Handlers) secretGroups(ctx context.Context) []view.SecretGroup {
	var key *keystore.Key
	if h.opts.Keys != nil {
		key = h.opts.Keys.EncryptKey(ctx, storage.ScopeInstance)
	}
	if h.opts.Workspaces == nil {
		return []view.SecretGroup{{Workspace: model.Default, Title: "Internal", Key: key}}
	}
	workspaces, _ := h.opts.Workspaces.List(ctx)
	groups := make([]view.SecretGroup, 0, len(workspaces))
	for _, ws := range workspaces {
		groups = append(groups, view.SecretGroup{Workspace: ws.Folder, Title: ws.Title, Key: key})
	}
	return groups
}

// workspaceOf returns the folder of the workspace a scope is in, "" for
// every workspace.
func (h *Handlers) workspaceOf(ctx context.Context, scope string) string {
	name, _ := h.opts.Secrets.Workspace(ctx, scope)
	return name
}

// scopeTitles names each scope the page lists: every workspace, then
// each workspace's.
func scopeTitles(groups []view.SecretGroup) map[string]string {
	titles := map[string]string{storage.ScopeAll: "All workspaces"}
	for _, group := range groups {
		titles[group.Shared()] = group.Title
	}
	return titles
}

// reachingWorkspace keeps the variables that reach a workspace: those of
// every workspace, its own, and its folders'.
func (h *Handlers) reachingWorkspace(ctx context.Context, secrets []model.Secret, workspace string) []model.Secret {
	result := make([]model.Secret, 0, len(secrets))
	for _, secret := range secrets {
		for _, term := range strings.Split(secret.Scope, ",") {
			term = strings.TrimSpace(term)
			if term == storage.ScopeAll || model.Canonical(h.workspaceOf(ctx, term)) == model.Canonical(workspace) {
				result = append(result, secret)
				break
			}
		}
	}
	return result
}

// integrationEnv returns what the integrations add to a run of the user,
// name and value, the value of a name that holds a credential redacted.
func (h *Handlers) integrationEnv(r *http.Request, p *view.Page) []view.EnvVar {
	if h.opts.Integrations == nil || p.User == nil {
		return nil
	}
	vars, err := h.opts.Integrations.Env(r.Context(), p.User.Name, p.User.EffectiveRole())
	if err != nil {
		return nil
	}
	result := make([]view.EnvVar, 0, len(vars))
	for _, name := range model.SortedKeys(vars) {
		item := view.EnvVar{Name: name, Value: vars[name]}
		if runner.Sensitive(name) {
			item.Value, item.Redacted = "", true
		}
		result = append(result, item)
	}
	return result
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
