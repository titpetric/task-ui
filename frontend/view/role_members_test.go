package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// Every role has an entry, empty or not.
func TestNewRoleMembers_Empty(t *testing.T) {
	t.Parallel()

	members := view.NewRoleMembers(nil, nil)
	assert.Equal(t, len(model.Roles()), len(members))
	for _, role := range model.Roles() {
		assert.Equal(t, 0, len(members[role].Users))
		assert.Equal(t, 0, len(members[role].Keys))
	}
}

// A user counts under the role they act with, the admin flag first.
func TestNewRoleMembers_Users(t *testing.T) {
	t.Parallel()

	members := view.NewRoleMembers([]model.User{
		{Name: "zed", Role: model.RoleRunner},
		{Name: "ann", Role: model.RoleRunner},
		{Name: "root", Role: model.RoleViewer, Admin: true},
		{Name: "guest", Role: model.RoleViewer},
	}, nil)
	assert.Equal(t, []string{"ann", "zed"}, members[model.RoleRunner].Users)
	assert.Equal(t, []string{"guest"}, members[model.RoleViewer].Users)
	assert.Equal(t, []string{"root"}, members[model.RoleAdmin].Users)
}

// A role holds the instance SSH keys linked to it; a workspace's key and
// an encryption key belong to no role.
func TestNewRoleMembers_Keys(t *testing.T) {
	t.Parallel()

	members := view.NewRoleMembers(nil, []keystore.Key{
		{Name: "deploy_ed25519", Purpose: keystore.PurposeSSH, Roles: []string{model.RoleRunner, model.RoleAdmin}},
		{Name: "backup_ed25519", Purpose: keystore.PurposeSSH, Roles: []string{model.RoleAdmin}},
		{Name: "workspace_acme_ci_ed25519", Purpose: keystore.PurposeSSH, Scope: keystore.ScopeWorkspace("acme"), Roles: []string{model.RoleRunner}},
		{Name: "task_encryption_ed25519", Purpose: keystore.PurposeEncrypt, Roles: []string{model.RoleAdmin}},
		{Name: "user_ops_ed25519", Purpose: keystore.PurposeSSH, Users: []string{"ops"}},
	})
	assert.Equal(t, []string{"backup_ed25519", "deploy_ed25519"}, members[model.RoleAdmin].Keys)
	assert.Equal(t, []string{"deploy_ed25519"}, members[model.RoleRunner].Keys)
	assert.Equal(t, 0, len(members[model.RoleViewer].Keys))
}
