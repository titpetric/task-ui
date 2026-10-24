package view

import (
	"strings"

	"github.com/titpetric/task-ui/folder"
)

// FolderMenu is the [...] menu in the head of a folder page: Pin or
// Unpin, and for a workspace or a project a link to its settings, for a
// workspace its SSH keys and secrets, Pull now, and for a project with a
// checkout Remove checkout, each behind its permission in the folder
// template.
type FolderMenu struct {
	// ID is the folder's: _ for the root.
	ID string
	// Workspace marks a workspace's page, Project a project's.
	Workspace bool
	Project   bool
	// Pull is the pull job of the git skill there, nil without one.
	Pull *folder.Job
}

// Kind names what the page is a folder of, for Pin: workspace, project or folder.
func (m *FolderMenu) Kind() string {
	switch {
	case m.Workspace:
		return "workspace"
	case m.Project:
		return "project"
	}
	return "folder"
}

// Settings returns the path of the settings page of the workspace or
// the project, "" for neither.
func (m *FolderMenu) Settings() string {
	switch {
	case m == nil:
		return ""
	case m.Workspace:
		return "/settings/workspaces/" + m.ID
	case m.Project:
		return "/settings/projects/" + m.ID
	}
	return ""
}

// KeysPage returns the path of the page of a workspace's SSH keys, ""
// for a project.
func (m *FolderMenu) KeysPage() string {
	if m == nil || !m.Workspace {
		return ""
	}
	return "/settings/workspaces/" + m.ID + "/keys"
}

// SecretsPage returns the path of the page of a workspace's secrets, ""
// for a project.
func (m *FolderMenu) SecretsPage() string {
	if m == nil || !m.Workspace {
		return ""
	}
	return "/settings/workspaces/" + m.ID + "/secrets"
}

// RemoveConfirm returns what Remove checkout asks before it removes the
// folder id and the checkout of remote.
func RemoveConfirm(id, remote string) string {
	if remote == "" {
		return "Remove the folder " + id + " and its checkout? The checkout has no remote to clone it again from."
	}
	return "Remove the folder " + id + " and its checkout? You can clone it again from " + remote + "."
}

// DeleteWorkspaceConfirm returns what Delete of a workspace asks, naming its nested projects.
func DeleteWorkspaceConfirm(folder string, nested []string) string {
	text := "Delete the workspace " + folder + " with its keys and secrets"
	if len(nested) > 0 {
		text += ", and the projects inside it with their checkouts: " + strings.Join(nested, ", ")
	}
	return text + "? Its directory goes once nothing else is in it."
}
