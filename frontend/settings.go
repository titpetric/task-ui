package frontend

import (
	"fmt"
	"net/http"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/model"
)

// settings is the index of what admins manage, the steps in the order
// an admin sets up an instance: for a user who may see any of them.
func (h *Handlers) settings(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, "Settings")
	if !p.CanSettings() {
		h.refuse(w, r)
		return
	}
	p.Crumbs = view.Crumbs("Settings")
	p.StepCounts = map[string]string{}
	// Each step shows how much it holds of what the user may see.
	if p.Can("key/view") {
		p.Keys, p.KeysDir = h.accessKeys(h.opts.Keys), h.opts.Keys.Dir
		keys := 0
		for _, key := range p.Keys {
			if !key.Backup {
				keys++
			}
		}
		p.StepCounts[view.StepKeys] = count(keys, "key", "keys")
	}
	if p.Can("secret/view") {
		secrets, _ := h.opts.Secrets.List(r.Context())
		p.StepCounts[view.StepSecrets] = count(len(userSecrets(secrets)), "secret", "secrets")
	}
	if p.Can("user/view") {
		users, _ := h.opts.Users.List(r.Context())
		p.StepCounts[view.StepUsers] = count(len(users), "user", "users")
		p.StepCounts[view.StepPermissions] = count(len(model.Roles()), "role", "roles")
	}
	if p.Can("workflow/view") {
		p.StepCounts[view.StepWorkflows] = count(len(p.SideWorkflows), "workflow", "workflows")
	}
	if p.Can("workspace/view") {
		p.StepCounts[view.StepWorkspaces] = count(len(h.workspaceList(r.Context())), "workspace", "workspaces")
	}
	if p.Can("project/view") {
		entries := folder.Entries(h.opts.Root, h.opts.Git, h.workspaceFolders(r.Context()))
		p.StepCounts[view.StepProjects] = count(len(entries), "folder", "folders")
	}
	h.render(w, http.StatusOK, "settings", p)
}

// instance shows what the server runs with, read only: for a user who
// may see any step of the settings.
func (h *Handlers) instance(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, "Instance")
	if !p.CanSettings() {
		h.refuse(w, r)
		return
	}
	p.Crumbs = view.Crumbs("Settings", "/settings", "Instance")
	p.Step = view.StepInstance
	instance := view.Instance{}
	if h.opts.Instance != nil {
		instance = *h.opts.Instance
	}
	if h.opts.Integrations != nil {
		values, _ := h.opts.Integrations.Values(r.Context(), integration.Beszel)
		instance.Beszel = values["url"] != "" && values["token"] != ""
	}
	p.Instance = &instance
	h.render(w, http.StatusOK, "instance", p)
}

// count says how many of a thing there are: 1 key, 2 keys.
func count(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}
