package view

import (
	"strconv"
	"strings"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
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

// DeleteBlocked returns the menu line of a workspace that holds secrets: delete them first.
func DeleteBlocked(secrets int) string {
	if secrets == 1 {
		return "Delete its 1 secret first"
	}
	return "Delete its " + strconv.Itoa(secrets) + " secrets first"
}

// DeleteWorkspaceConfirm returns what Delete of a workspace asks, or
// what Empty of the default workspace asks.
func DeleteWorkspaceConfirm(folder string, nested []string) string {
	projects := ""
	if len(nested) > 0 {
		projects = ", and the projects inside it with their checkouts: " + strings.Join(nested, ", ")
	}
	if folder == model.Default {
		return "Empty the default workspace " + folder + "? This deletes its home checkout, its keys and secrets" + projects + ". The root stays, with an empty home and src, as do the other workspaces."
	}
	return "Delete the workspace " + folder + " with its keys and secrets" + projects + "? Its directory goes once nothing else is in it."
}
