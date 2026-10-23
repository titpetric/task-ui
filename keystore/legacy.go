package keystore

import (
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Legacy is where the older layout kept the keys: the instance's
// encryption key under Data/instance/encrypt, each workspace's under
// Data/workspaces/<id>/encrypt and ssh, the instance SSH keys as id_*
// files in the key folder itself, and who got which in the policy
// tables.
type Legacy struct {
	Data       string
	Workspaces []model.Workspace
	Policies   *storage.SSHKeyPolicyStorage
	Grants     *storage.WorkspaceKeyGrantStorage
}
