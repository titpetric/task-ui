package view

import (
	"net/url"
)

// The pages that add a thing, and the pages of the things a workspace
// owns. Every template and menu that links "add" reads them here, so
// the folder menu, the row menus, the dashboard and the empty states
// cannot link different places for the same action.
const (
	newWorkspacePath = "/settings/workspaces/new"
	newProjectPath   = "/settings/projects/new"
	newUserPath      = "/settings/users/new"
	newSecretPath    = "/settings/secrets/new"
	newWorkflowPath  = "/workflows/new"
	keysPath         = "/settings/keys"
)

// AddProjectHref returns the form that clones a project, the workspace
// chosen; "" leaves the choice to the form.
func AddProjectHref(workspace string) string {
	if workspace == "" {
		return newProjectPath
	}
	return newProjectPath + "?workspace=" + url.QueryEscape(workspace)
}

// WorkspaceHref returns the settings page of a workspace.
func WorkspaceHref(workspace string) string {
	return "/settings/workspaces/" + workspace
}

// WorkspaceKeysHref returns the page of a workspace's SSH keys.
func WorkspaceKeysHref(workspace string) string {
	return WorkspaceHref(workspace) + "/keys"
}

// WorkspaceSecretsHref returns the secrets page filtered to what
// reaches a workspace.
func WorkspaceSecretsHref(workspace string) string {
	return "/settings/secrets?workspace=" + url.QueryEscape(workspace)
}

// UserKeyHref returns the form that adds an SSH key linked to a user:
// kind is "generate" or "upload".
func UserKeyHref(kind, user string) string {
	return keysPath + "/" + kind + "?for=" + url.QueryEscape(user)
}

// CreateItems returns what the user may add anywhere, in the order of
// the setup: a workspace, a project, a user, an SSH key, a variable and
// a workflow, each behind the rule of the page that adds it. The
// dashboard's Add menu draws them.
func (p *Page) CreateItems() []MenuItem {
	var items []MenuItem
	add := func(allowed bool, icon, label, href string) {
		if !allowed {
			return
		}
		item, _ := NewMenuItem(icon, label, "href", href)
		items = append(items, item)
	}
	add(p.Can("workspace/edit"), "folder-open", "Add workspace", newWorkspacePath)
	add(p.Can("project/edit"), "folder", "Add project", newProjectPath)
	add(p.CanAddUsers && p.Can("user/edit"), "users", "Add user", newUserPath)
	add(p.Can("key/edit"), "key", "Generate key", keysPath+"/generate")
	add(p.Can("key/edit"), "upload", "Upload key", keysPath+"/upload")
	add(p.Can("secret/edit"), "secret", "Add variable", newSecretPath)
	add(p.Can("workflow/edit"), "flow", "Add workflow", newWorkflowPath)
	return items
}
