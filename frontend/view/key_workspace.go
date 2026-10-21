package view

import (
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
)

// KeyWorkspace is a workspace on the keys page: its projects, how many
// secrets it stores, its one encryption key, nil while it has none, and
// its SSH keys, backups left out.
type KeyWorkspace struct {
	Workspace model.Workspace
	Projects  []string
	Secrets   int
	Encrypt   *sshkey.Key
	SSH       []sshkey.Key
}

// Names returns the names of the workspace's SSH keys.
func (k KeyWorkspace) Names() []string {
	names := make([]string, 0, len(k.SSH))
	for _, key := range k.SSH {
		names = append(names, key.Name)
	}
	return names
}
