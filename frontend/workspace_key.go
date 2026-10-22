package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
)

// keyWorkspace returns the workspace a key page names, answering the
// request when there is none or the keyring is off. The page checks the
// permission first.
func (h *Handlers) keyWorkspace(w http.ResponseWriter, r *http.Request) *model.Workspace {
	name := param(r, "folder")
	if h.opts.Workspaces == nil || h.opts.WorkspaceKeys == nil {
		h.missing(w, r, "No workspace named "+name+".")
		return nil
	}
	workspace, err := h.opts.Workspaces.GetByFolder(r.Context(), name)
	if err != nil || workspace == nil {
		h.missing(w, r, "No workspace named "+name+".")
		return nil
	}
	return workspace
}

// workspaceKeys puts a workspace's keys on its page: the encryption
// key, the SSH keys with their backups, and whom each goes to.
// Without the keyring the page has none.
func (h *Handlers) workspaceKeys(r *http.Request, p *view.Page) {
	p.KeyPolicies, p.KeyUsers = map[string][]string{}, map[string][]string{}
	if h.opts.WorkspaceKeys == nil || p.Workspace == nil {
		return
	}
	id := p.Workspace.ID
	p.EncryptKey = h.opts.WorkspaceKeys.EncryptKey(id)
	p.Keys, p.KeysDir = h.opts.WorkspaceKeys.SSH(id).List(), h.opts.WorkspaceKeys.Dir(id)
	if h.opts.KeyGrants != nil {
		if roles, users, err := h.opts.KeyGrants.All(r.Context(), id); err == nil {
			p.KeyPolicies, p.KeyUsers = roles, users
		}
	}
	p.Users, _ = h.opts.Users.List(r.Context())
}

// workspaceKeyPage starts a page of a workspace's keys.
func (h *Handlers) workspaceKeyPage(r *http.Request, workspace *model.Workspace, title, crumb string) *view.Page {
	p := h.page(r, title)
	p.Crumbs = view.Crumbs("Settings", "/settings", "Workspaces", "/settings/workspaces", workspace.Folder, "/settings/workspaces/"+workspace.Folder, crumb)
	p.Workspace = workspace
	h.workspaceKeys(r, p)
	return p
}

// workspaceKeyGenerate shows the form that generates an SSH key for a
// workspace.
func (h *Handlers) workspaceKeyGenerate(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	workspace := h.keyWorkspace(w, r)
	if workspace == nil {
		return
	}
	h.render(w, http.StatusOK, "workspace_key_generate", h.workspaceKeyPage(r, workspace, "Generate key", "Generate key"))
}

// workspaceKeyUpload shows the form that uploads an SSH key for a
// workspace, or with ?encrypt=1 its encryption key.
func (h *Handlers) workspaceKeyUpload(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	workspace := h.keyWorkspace(w, r)
	if workspace == nil {
		return
	}
	title := "Upload key"
	encrypt := r.URL.Query().Get("encrypt") == "1"
	if encrypt {
		title = "Upload encryption key"
	}
	p := h.workspaceKeyPage(r, workspace, title, title)
	p.Form = map[string]string{}
	if encrypt {
		p.Form["encrypt"] = "true"
	}
	h.render(w, http.StatusOK, "workspace_key_upload", p)
}

// workspaceKeyEdit shows the form that sets whose runs get one of a
// workspace's SSH keys.
func (h *Handlers) workspaceKeyEdit(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/view") {
		return
	}
	workspace := h.keyWorkspace(w, r)
	if workspace == nil {
		return
	}
	name := param(r, "name")
	for _, key := range h.opts.WorkspaceKeys.SSH(workspace.ID).List() {
		if key.Name != name || key.Backup {
			continue
		}
		p := h.workspaceKeyPage(r, workspace, key.Name, key.Name)
		p.Key = &key
		h.render(w, http.StatusOK, "workspace_key_edit", p)
		return
	}
	h.notFound(w, r)
}
