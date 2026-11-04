package view_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
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

// roleBlock returns the block of a role on the permissions page.
func roleBlock(t *testing.T, body, role string) string {
	t.Helper()
	at := strings.Index(body, `id="role-`+role+`"`)
	assert.True(t, at >= 0, "a block for "+role)
	rest := body[at:]
	return rest[:strings.Index(rest, "</section>")]
}

// Each role's block lists the users who act with it and the instance SSH
// keys linked to it, so a role leads to both.
func TestPermissions_RoleMembers(t *testing.T) {
	t.Parallel()

	p := admin(t)
	p.CanAddUsers = true
	p.Matrices = []view.Matrix{{ID: "role-runner", Title: "runner"}, {ID: "role-viewer", Title: "viewer"}}
	p.RoleMembers = map[string]view.RoleMembers{
		"runner": {Users: []string{"ann", "zed"}, Keys: []string{"deploy_ed25519"}},
	}
	body := render(t, "permissions", p)

	runner := roleBlock(t, body, "runner")
	assert.Contains(t, runner, `<a class="key-chip" href="/users/ann">ann</a>`)
	assert.Contains(t, runner, `<a class="key-chip" href="/users/zed">zed</a>`)
	assert.Contains(t, runner, `<a class="key-chip" href="/settings/keys/deploy_ed25519">deploy_ed25519</a>`)
	assert.Contains(t, runner, `href="/settings/users/new">Add user</a>`)
	assert.Contains(t, runner, `href="/settings/keys/generate">Generate key</a>`)

	viewer := roleBlock(t, body, "viewer")
	assert.Equal(t, 2, strings.Count(viewer, `<span class="faint">none</span>`), "no user and no key")
}

// The keys and the add links of a role follow the user's rules.
func TestPermissions_RoleMembersRules(t *testing.T) {
	t.Parallel()

	p := admin(t)
	p.Matrices = []view.Matrix{{ID: "role-runner", Title: "runner"}}
	p.RoleMembers = map[string]view.RoleMembers{"runner": {Users: []string{"ann"}, Keys: []string{"deploy_ed25519"}}}
	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"user/view": permission.Allow})
	runner := roleBlock(t, render(t, "permissions", p), "runner")
	assert.Contains(t, runner, `href="/users/ann"`)
	assert.NotContains(t, runner, "deploy_ed25519", "the keys take key/view")
	assert.NotContains(t, runner, "Add user", "adding a user takes user/edit")
	assert.NotContains(t, runner, "Generate key")

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"user/view": permission.Allow, "key/view": permission.Allow})
	runner = roleBlock(t, render(t, "permissions", p), "runner")
	assert.Contains(t, runner, `href="/settings/keys/deploy_ed25519"`)
	assert.NotContains(t, runner, "Generate key", "adding a key takes key/edit")
}

// The role on the users list links to what the role may do.
func TestUsers_RoleLinksToPermissions(t *testing.T) {
	t.Parallel()

	p := admin(t)
	p.Users = []model.User{{ID: "u1", Name: "ops", Role: model.RoleRunner}, {ID: "u2", Name: "root", Role: model.RoleViewer, Admin: true}}
	body := render(t, "users", p)
	assert.Contains(t, body, `<a class="badge" href="/settings/permissions#role-runner" title="What runner may do">runner</a>`)
	assert.Contains(t, body, `<a class="badge signal" href="/settings/permissions#role-admin" title="What admin may do">admin</a>`)
}
