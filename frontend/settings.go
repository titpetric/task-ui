package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
)

// settings is the index of what admins manage: for a user who may see
// any of its sections.
func (h *Handlers) settings(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, "Settings")
	if !p.CanSettings() {
		h.refuse(w, r)
		return
	}
	p.Crumbs = view.Crumbs("Settings")
	// Each section shows what the user may see of it.
	if p.Can("project/view") {
		p.Entries = folder.Entries(h.opts.Root, h.opts.Git, h.workspaceFolders(r.Context()))
	}
	if p.Can("workspace/view") {
		p.WorkspaceList = h.workspaceList(r.Context())
	}
	if p.Can("key/view") {
		p.Keys, p.KeysDir = h.accessKeys(h.opts.Keys), h.opts.Keys.Dir
	}
	if p.Can("secret/view") {
		secrets, _ := h.opts.Secrets.List(r.Context())
		p.Secrets = userSecrets(secrets)
	}
	if p.Can("user/view") {
		p.Users, _ = h.opts.Users.List(r.Context())
	}
	h.render(w, http.StatusOK, "settings", p)
}
