package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
)

// settings is the index of what admins manage.
func (h *Handlers) settings(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	p := h.page(r, "Settings")
	p.Crumbs = view.Crumbs("Settings")
	p.Entries = folder.Entries(h.opts.Root, h.opts.Git, h.workspaceFolders(r.Context()))
	p.WorkspaceList = h.workspaceList(r.Context())
	p.Keys, p.KeysDir = h.opts.Keys.List(), h.opts.Keys.Dir
	secrets, _ := h.opts.Secrets.List(r.Context())
	p.Secrets = userSecrets(secrets)
	p.Users, _ = h.opts.Users.List(r.Context())
	h.render(w, http.StatusOK, "settings", p)
}
