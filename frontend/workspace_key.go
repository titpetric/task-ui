package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// keyWorkspace returns the workspace a key page names, answering the
// request when there is none or the keyring is off. The page checks the
// permission first.
func (h *Handlers) keyWorkspace(w http.ResponseWriter, r *http.Request) *model.Workspace {
	name := param(r, "folder")
	if h.opts.Workspaces == nil || h.opts.Keys == nil {
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

// workspaceKeys puts a workspace's SSH keys on its page. Without the
// key folder the page has none.
//
// Whom each key goes to is not a question the page asks: a key of this
// workspace goes to every run of this workspace, so there are no roles
// and no users to fill in beside it.
func (h *Handlers) workspaceKeys(r *http.Request, p *view.Page) {
	if h.opts.Keys == nil || p.Workspace == nil {
		return
	}
	scope := keystore.ScopeWorkspace(p.Workspace.ID)
	p.Keys, _ = h.opts.Keys.Of(r.Context(), scope, keystore.PurposeSSH)
	p.KeysDir = h.opts.Keys.Dir
}

// workspaceKeyPage starts a page of a workspace's keys.
func (h *Handlers) workspaceKeyPage(r *http.Request, workspace *model.Workspace, title string, crumbs ...string) *view.Page {
	p := h.page(r, title)
	p.Crumbs = view.Crumbs(append([]string{"Settings", "/settings", "Workspaces", "/settings/workspaces", crumbName(workspace.Folder), "/settings/workspaces/" + workspace.Folder}, crumbs...)...)
	p.Step = view.StepWorkspaces
	p.Workspace = workspace
	h.workspaceKeys(r, p)
	return p
}

// workspaceKeyList lists a workspace's SSH keys, with Generate key,
// Upload key and Edit links in dialogs on the page.
func (h *Handlers) workspaceKeyList(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/view") {
		return
	}
	workspace := h.keyWorkspace(w, r)
	if workspace == nil {
		return
	}
	h.render(w, http.StatusOK, "workspace_keys", h.workspaceKeyPage(r, workspace, "Keys of "+workspace.Folder, "Keys"))
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
// workspace.
func (h *Handlers) workspaceKeyUpload(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "workspace/edit") {
		return
	}
	workspace := h.keyWorkspace(w, r)
	if workspace == nil {
		return
	}
	p := h.workspaceKeyPage(r, workspace, "Upload key", "Upload key")
	p.Form = map[string]string{}
	h.render(w, http.StatusOK, "workspace_key_upload", p)
}

