package workspace

import (
	"strings"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Format is the version of the manifest a bundle holds.
const Format = 1

// Manifest is what a bundle holds beside the keys: the workspace's
// row, its projects, the grants of its SSH keys, its env files and its
// secrets. Folders are relative to the workspace, "" its own folder; a
// secret's scope is a relative folder, or workspace: for the secrets it
// shares. Secrets go plain, inside the bundle's passphrase, and the
// importing instance encrypts them for its own encryption key; an
// older bundle's carry ciphertext and the workspace's encryption key
// that opens them. IDs and dates stay behind; the importing instance
// makes its own.
type Manifest struct {
	Format int `json:"format"`
	// Version is the task-ui that wrote the bundle.
	Version   string                    `json:"version"`
	Workspace model.Workspace           `json:"workspace"`
	Projects  []model.Project           `json:"projects"`
	Grants    []model.WorkspaceKeyGrant `json:"grants"`
	EnvFiles  []model.EnvFile           `json:"env_files"`
	Secrets   []model.Secret            `json:"secrets"`
}

// Contains reports whether a folder belongs to a workspace: the
// workspace's folder and those inside it, or for the default workspace
// every folder outside the named ones and task-ui's own.
func Contains(workspace string, named []string, folder string) bool {
	top, _, _ := strings.Cut(folder, "/")
	if workspace != model.Default {
		return top == workspace
	}
	if model.IsReserved(folder) {
		return false
	}
	for _, name := range named {
		if name == top {
			return false
		}
	}
	return true
}

// Relative returns a folder of a workspace relative to it: "" for the
// workspace's own folder. The default workspace's folders are relative
// to the root already.
func Relative(workspace, folder string) string {
	if folder == workspace {
		return ""
	}
	if workspace == model.Default {
		return folder
	}
	return strings.TrimPrefix(folder, workspace+"/")
}

// Absolute returns the folder a relative one names in a workspace.
func Absolute(workspace, relative string) string {
	if relative == "" {
		return workspace
	}
	if workspace == model.Default {
		return relative
	}
	return workspace + "/" + relative
}

// RelativeScope returns a secret's scope relative to its workspace:
// workspace: for its shared secrets, else the relative folder.
func RelativeScope(workspace, scope string) string {
	if strings.HasPrefix(scope, storage.ScopeWorkspace) {
		return storage.ScopeWorkspace
	}
	return Relative(workspace, scope)
}

// AbsoluteScope returns the scope a relative one names in a workspace.
func AbsoluteScope(workspace, relative string) string {
	if relative == storage.ScopeWorkspace {
		return storage.SharedScope(workspace)
	}
	return Absolute(workspace, relative)
}
