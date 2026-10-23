package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// A role row lists the instance SSH keys of the role; a user row the
// keys linked to them by name and, apart, those their role gets. An
// encryption key and a workspace's key are nobody's here.
func TestKeyHolder(t *testing.T) {
	t.Parallel()
	keys := []keystore.Key{
		{Name: "deploy_rsa4096", Purpose: keystore.PurposeSSH, Scope: "instance", Roles: []string{"admin"}},
		{Name: "user_alice_ed25519", Purpose: keystore.PurposeSSH, Scope: "instance", Users: []string{"alice"}},
		{Name: "task_encryption_ed25519", Purpose: keystore.PurposeEncrypt, Scope: "instance", Roles: []string{"admin"}},
		{Name: "workspace_malta_ssh_ed25519", Purpose: keystore.PurposeSSH, Scope: keystore.ScopeWorkspace("w1"), Roles: []string{"admin"}},
	}
	users := []model.User{{Name: "alice", Role: model.RoleAdmin}, {Name: "bob", Role: model.RoleViewer}}

	var rows []view.KeyHolder = view.KeyHolders(keys, users)
	assert.Len(t, rows, len(model.Roles())+2)
	assert.Equal(t, "admin", rows[0].Role)
	assert.Len(t, rows[0].Direct, 1)
	alice := rows[len(model.Roles())]
	assert.Equal(t, "alice", alice.User.Name)
	assert.Equal(t, "user_alice_ed25519", alice.Direct[0].Name)
	assert.Equal(t, "deploy_rsa4096", alice.ViaRole[0].Name)
	bob := rows[len(model.Roles())+1]
	assert.Len(t, bob.Direct, 0)
	assert.Len(t, bob.ViaRole, 0)
}
