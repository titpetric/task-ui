package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// UserKeys names the instance SSH keys each user gets, by their name
// and by their effective role, sorted; an encryption key, a workspace's
// key and a key linked to nobody they know count for no one, and a user
// with no key has no entry.
func TestUserKeys(t *testing.T) {
	t.Parallel()
	keys := []keystore.Key{
		{Name: "forge_ed25519", Purpose: keystore.PurposeSSH, Scope: "instance", Roles: []string{"runner"}, Users: []string{"admin"}},
		{Name: "deploy_rsa4096", Purpose: keystore.PurposeSSH, Scope: "instance", Roles: []string{"admin"}},
		{Name: "user_alice_ed25519", Purpose: keystore.PurposeSSH, Scope: "instance", Users: []string{"alice"}},
		{Name: "spare_ed25519", Purpose: keystore.PurposeSSH, Scope: "instance"},
		{Name: "task_encryption_ed25519", Purpose: keystore.PurposeEncrypt, Scope: "instance", Roles: []string{"admin"}},
		{Name: "workspace_malta_ssh_ed25519", Purpose: keystore.PurposeSSH, Scope: keystore.ScopeWorkspace("w1"), Roles: []string{"admin"}},
	}
	users := []model.User{
		{Name: "admin", Role: model.RoleAdmin},
		{Name: "alice", Role: model.RoleRunner},
		{Name: "ops", Role: model.RoleViewer, Admin: true},
		{Name: "guest", Role: model.RoleViewer},
	}
	got := view.UserKeys(keys, users)
	assert.Equal(t, []string{"deploy_rsa4096", "forge_ed25519"}, got["admin"], "by role and by name, sorted")
	assert.Equal(t, []string{"forge_ed25519", "user_alice_ed25519"}, got["alice"])
	assert.Equal(t, []string{"deploy_rsa4096"}, got["ops"], "the effective role, admin, decides")
	_, has := got["guest"]
	assert.False(t, has, "no entry without a key")
	assert.Len(t, view.UserKeys(nil, users), 0)
}
