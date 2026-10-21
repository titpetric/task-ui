package view

import (
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
)

// SecretGroup is a workspace on the secrets pages: the secrets its
// key encrypts, those it shares with its folders first, and the folders
// a secret of its can name as its scope.
type SecretGroup struct {
	// Workspace is the workspace's folder, _ for the default one.
	Workspace string
	Title     string
	// Key is the workspace's encryption key, nil while it has none.
	Key     *sshkey.Key
	Secrets []model.Secret
	Folders []string
}

// Shared returns the scope of the secrets the workspace shares with its
// folders.
func (g SecretGroup) Shared() string {
	return "workspace:" + g.Workspace
}
