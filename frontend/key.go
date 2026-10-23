package frontend

import (
	"context"
	"net/http"
	"path/filepath"
	"slices"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// keys lists every key of the key folder: the instance's and each
// workspace's encryption key, which stores secrets, and the SSH keys,
// which connect to hosts, with who gets each and the one git pulls with.
func (h *Handlers) keys(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "key/view") {
		return
	}
	p := h.page(r, "Keys")
	p.Crumbs = view.Crumbs("Settings", "/settings", "Keys")
	p.Step = view.StepKeys
	h.fillKeys(r.Context(), p)
	p.KeyFolders = h.keyFolders()
	users, _ := h.opts.Users.List(r.Context())
	p.KeyHolders = view.KeyHolders(p.Keys, users)
	p.KeyWorkspaces = h.keyWorkspaces(r.Context(), p.Folders)
	h.render(w, http.StatusOK, "keys", p)
}

// fillKeys puts the key folder on a page: every key, the links of each
// by name, the git key and the instance's encryption key.
func (h *Handlers) fillKeys(ctx context.Context, p *view.Page) {
	p.KeyPolicies, p.KeyUsers = map[string][]string{}, map[string][]string{}
	if h.opts.Keys == nil {
		return
	}
	p.Keys, _ = h.opts.Keys.List(ctx)
	p.KeysDir = h.opts.Keys.Dir
	for _, key := range p.Keys {
		p.KeyPolicies[key.Name], p.KeyUsers[key.Name] = key.Roles, key.Users
		if key.Git {
			p.GitKey = key.Name
		}
	}
	p.InstanceKey = h.opts.Keys.EncryptKey(ctx, storage.ScopeInstance)
	if h.opts.KeyReads != nil {
		p.KeyReads, _ = h.opts.KeyReads.Counts(ctx)
		p.KeyLastRead, _ = h.opts.KeyReads.Last(ctx)
	}
}

// keyWorkspaces returns a row per workspace, the default one first: its
// projects, the count of its secrets, its encryption key and its SSH
// keys.
func (h *Handlers) keyWorkspaces(ctx context.Context, folders []*folder.Folder) []view.KeyWorkspace {
	if h.opts.Workspaces == nil || h.opts.Keys == nil {
		return nil
	}
	workspaces, err := h.opts.Workspaces.List(ctx)
	if err != nil {
		return nil
	}
	rows := make([]view.KeyWorkspace, 0, len(workspaces))
	index := map[string]int{}
	for _, ws := range workspaces {
		index[ws.Folder] = len(rows)
		scope := keystore.ScopeWorkspace(ws.ID)
		row := view.KeyWorkspace{Workspace: ws, Encrypt: h.opts.Keys.EncryptKey(ctx, scope)}
		row.SSH, _ = h.opts.Keys.Of(ctx, scope, keystore.PurposeSSH)
		rows = append(rows, row)
	}
	for _, f := range folders {
		name := h.workspaceOf(ctx, f.ID)
		i, ok := index[name]
		if !ok || !project(f, name, index) {
			continue
		}
		rows[i].Projects = append(rows[i].Projects, f.ID)
	}
	if h.opts.Secrets != nil {
		secrets, _ := h.opts.Secrets.List(ctx)
		for _, secret := range userSecrets(secrets) {
			if i, ok := index[h.workspaceOf(ctx, secret.Scope)]; ok {
				rows[i].Secrets++
			}
		}
	}
	return rows
}

// project reports whether a folder is a project of the workspace: one
// level inside a named workspace, or at the root, and no workspace, for
// the default one.
func project(f *folder.Folder, workspace string, workspaces map[string]int) bool {
	if workspace != model.Default {
		return f.Depth() == 1
	}
	_, named := workspaces[f.ID]
	return f.Depth() == 0 && !named
}

// instanceKey puts the instance's encryption key and the git key on a
// page.
func (h *Handlers) instanceKey(r *http.Request, p *view.Page) {
	h.fillKeys(r.Context(), p)
}

// keyGenerate shows the generate form.
func (h *Handlers) keyGenerate(w http.ResponseWriter, r *http.Request) {
	h.keyForm(w, r, "Generate key", "Generate", "key_generate")
}

// keyUpload shows the upload form.
func (h *Handlers) keyUpload(w http.ResponseWriter, r *http.Request) {
	h.keyForm(w, r, "Upload key", "Upload", "key_upload")
}

// keyForm shows a form that adds a key.
func (h *Handlers) keyForm(w http.ResponseWriter, r *http.Request, title, crumb, template string) {
	if !h.require(w, r, "key/edit") {
		return
	}
	p := h.page(r, title)
	p.Crumbs = view.Crumbs("Settings", "/settings", "Keys", "/settings/keys", crumb)
	p.Step = view.StepKeys
	if h.opts.Keys != nil {
		p.KeysDir = h.opts.Keys.Dir
	}
	p.Users, _ = h.opts.Users.List(r.Context())
	// ?for=<user> starts a key of a user: user_<name>, linked to them.
	p.Form = map[string]string{}
	if user := r.URL.Query().Get("for"); user != "" {
		p.Form["name"], p.Form["for"] = "user_"+user, user
	}
	h.render(w, http.StatusOK, template, p)
}

// keyFolders lists the .ssh/ folders of the projects, the checkouts at
// the top of the root, that hold a private key. Runs do not load them, so
// the page warns: an admin uploads a key a project needs and links it.
func (h *Handlers) keyFolders() []view.KeyFolder {
	var result []view.KeyFolder
	for _, f := range h.opts.Catalogue.List() {
		if f.Depth() != 0 || h.opts.Git == nil || !h.opts.Git.IsRepo(f.Dir) {
			continue
		}
		dir := filepath.Join(f.Dir, ".ssh")
		keys := (&sshkey.Store{Dir: dir}).List()
		if len(keys) == 0 {
			continue
		}
		result = append(result, view.KeyFolder{Project: f.ID, Dir: dir, Keys: keys})
	}
	return result
}

// keyEdit shows the form that sets whose runs get an SSH key and
// whether git pulls with it.
func (h *Handlers) keyEdit(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "key/view") || h.opts.Keys == nil {
		return
	}
	key, err := h.opts.Keys.Get(r.Context(), param(r, "name"))
	if err != nil {
		h.notFound(w, r)
		return
	}
	p := h.page(r, key.Name)
	p.Crumbs = view.Crumbs("Settings", "/settings", "Keys", "/settings/keys", key.Name)
	p.Step = view.StepKeys
	h.fillKeys(r.Context(), p)
	p.Key = key
	p.Users, _ = h.opts.Users.List(r.Context())
	h.render(w, http.StatusOK, "key_edit", p)
}

// myKeys lists the instance SSH keys the signed in user's runs get, by
// role or by name, to copy their public halves. Every user has it.
func (h *Handlers) myKeys(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, "Your SSH keys")
	p.Crumbs = view.Crumbs("Your SSH keys")
	p.KeyUsers = map[string][]string{}
	if h.opts.Keys != nil && p.User != nil {
		keys, _ := h.opts.Keys.List(r.Context())
		for _, key := range keys {
			if key.Purpose == keystore.PurposeSSH && key.Workspace() == "" && (slices.Contains(key.Roles, p.User.EffectiveRole()) || slices.Contains(key.Users, p.User.Name)) {
				p.Keys = append(p.Keys, key)
				p.KeyUsers[key.Name] = key.Users
			}
		}
	}
	h.render(w, http.StatusOK, "my_keys", p)
}

// userKeys returns the instance SSH keys linked to a user by name.
func (h *Handlers) userKeys(ctx context.Context, user *model.User) []keystore.Key {
	if h.opts.Keys == nil {
		return nil
	}
	keys, _ := h.opts.Keys.List(ctx)
	var result []keystore.Key
	for _, key := range keys {
		if slices.Contains(key.Users, user.Name) {
			result = append(result, key)
		}
	}
	return result
}
