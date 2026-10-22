package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
	"github.com/titpetric/task-ui/storage"
)

// The migration seeds the roles with what they do without policies, and
// claude with notes, prompts and runs.
func TestPermissionPolicyStorage_Policy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	policies := storage.NewPermissionPolicyStorage(newDB(t))

	runner, err := policies.Policy(ctx, permission.KindRole, model.RoleRunner)
	assert.NoError(t, err)
	assert.Equal(t, permission.Allow, runner["run"])
	assert.Equal(t, permission.Allow, runner["note/edit"])
	assert.Equal(t, "", runner["prompt/view"])
	assert.Equal(t, "", runner["secret/view"])

	viewer, err := policies.Policy(ctx, permission.KindRole, model.RoleViewer)
	assert.NoError(t, err)
	assert.Equal(t, "", viewer["run"])
	assert.Equal(t, permission.Allow, viewer["note/view"])
	assert.Equal(t, permission.Allow, viewer["workflow/edit"])

	admin, err := policies.Policy(ctx, permission.KindRole, model.RoleAdmin)
	assert.NoError(t, err)
	assert.Len(t, admin, 0)

	claude, err := policies.Policy(ctx, permission.KindUser, "claude")
	assert.NoError(t, err)
	assert.Equal(t, permission.Policy{"note/view": "allow", "note/edit": "allow", "prompt/view": "allow", "prompt/edit": "allow", "run": "allow"}, claude)
}

func TestPermissionPolicyStorage_Put(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	policies := storage.NewPermissionPolicyStorage(newDB(t))

	assert.NoError(t, policies.Put(ctx, permission.KindUser, "ops", "secret/all", permission.Allow))
	assert.NoError(t, policies.Put(ctx, permission.KindUser, "ops", "secret/delete", permission.Deny))
	ops, err := policies.Policy(ctx, permission.KindUser, "ops")
	assert.NoError(t, err)
	assert.Equal(t, permission.Policy{"secret/all": "allow", "secret/delete": "deny"}, ops)

	assert.NoError(t, policies.Put(ctx, permission.KindUser, "ops", "secret/delete", permission.Allow), "a write replaces the entry")
	assert.NoError(t, policies.Put(ctx, permission.KindUser, "ops", "secret/all", ""), "none removes it")
	ops, err = policies.Policy(ctx, permission.KindUser, "ops")
	assert.NoError(t, err)
	assert.Equal(t, permission.Policy{"secret/delete": "allow"}, ops)

	assert.Error(t, policies.Put(ctx, "group", "ops", "run", permission.Allow))
	assert.Error(t, policies.Put(ctx, permission.KindUser, "", "run", permission.Allow))
	assert.Error(t, policies.Put(ctx, permission.KindRole, "root", "run", permission.Allow))
	assert.Error(t, policies.Put(ctx, permission.KindRole, model.RoleAdmin, "run", permission.Deny), "the admin role is fixed")
	assert.Error(t, policies.Put(ctx, permission.KindUser, "ops", "secret/write", permission.Allow))
	assert.Error(t, policies.Put(ctx, permission.KindUser, "ops", "run", "maybe"))
}

// A user's deny beats the role's allow, a user's allow adds to the role,
// and nobody may do what neither sets; an admin may do everything.
func TestPermissionPolicyStorage_For(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))
	policies := users.Policies()
	ops := &model.User{Name: "ops", Role: model.RoleRunner}

	assert.True(t, users.Permissions(ctx, ops).Allowed("run"))
	assert.False(t, users.Permissions(ctx, ops).Allowed("secret/view"), "unset is denied")

	assert.NoError(t, policies.Put(ctx, permission.KindUser, "ops", "run", permission.Deny))
	assert.False(t, users.Permissions(ctx, ops).Allowed("run"), "the user's deny beats the role's allow")
	assert.True(t, users.Permissions(ctx, &model.User{Name: "dev", Role: model.RoleRunner}).Allowed("run"))

	assert.NoError(t, policies.Put(ctx, permission.KindUser, "ops", "secret/view", permission.Allow))
	assert.True(t, users.Permissions(ctx, ops).Allowed("secret/view"))

	assert.NoError(t, policies.Put(ctx, permission.KindRole, model.RoleRunner, "secret/view", permission.Deny))
	assert.False(t, users.Permissions(ctx, ops).Allowed("secret/view"), "the role's deny beats the user's allow")

	root := &model.User{Name: "root", Role: model.RoleAdmin}
	assert.NoError(t, policies.Put(ctx, permission.KindUser, "root", "run", permission.Deny))
	assert.True(t, users.Permissions(ctx, root).Allowed("run"), "no deny reaches an admin")
	assert.False(t, users.Permissions(ctx, nil).Allowed("note/view"))
}

// Running a job takes the run permission, a folder the user sees and
// allow lists that name the user or the role: the seeded roles run as
// the roles ran before the policies.
func TestPermissionPolicyStorage_For_run(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))
	can := func(user *model.User, folder string, folderAllow, jobAllow []string) bool {
		return users.Permissions(ctx, user).Run(folder, folderAllow, jobAllow)
	}
	admin := &model.User{Name: "root", Role: model.RoleAdmin}
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	dev := &model.User{Name: "dev", Role: model.RoleRunner}
	scoped := &model.User{Name: "web", Role: model.RoleRunner, Folders: `["web"]`}
	guest := &model.User{Name: "guest", Role: model.RoleViewer}
	locked := []string{"ops"}
	admins := []string{"admin"}

	assert.True(t, can(admin, "infra", locked, admins))
	assert.True(t, can(ops, "infra", locked, nil))
	assert.False(t, can(ops, "infra", locked, admins))
	assert.True(t, can(dev, "web", nil, nil))
	assert.False(t, can(dev, "infra", locked, nil))
	assert.True(t, can(scoped, "web", nil, nil))
	assert.False(t, can(scoped, "infra", nil, nil))
	assert.False(t, can(guest, "web", nil, nil))
	assert.False(t, can(nil, "web", nil, nil))
	assert.True(t, can(ops, "infra", []string{"runner"}, nil), "a role in the allow list")

	assert.NoError(t, users.Policies().Put(ctx, permission.KindUser, "guest", "run", permission.Allow))
	assert.True(t, can(guest, "web", nil, nil), "a viewer allowed to run")
	assert.False(t, can(guest, "infra", locked, nil), "still as the allow lists say")
}

func TestPermissionPolicyStorage_DeleteSubject(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	policies := storage.NewPermissionPolicyStorage(newDB(t))
	assert.NoError(t, policies.Put(ctx, permission.KindUser, "ops", "note/view", permission.Allow))
	assert.NoError(t, policies.Put(ctx, permission.KindUser, "dev", "note/view", permission.Allow))
	policy, err := policies.Policy(ctx, permission.KindUser, "ops")
	assert.NoError(t, err)
	assert.Len(t, policy, 1)

	assert.NoError(t, policies.DeleteSubject(ctx, permission.KindUser, "ops"))
	policy, err = policies.Policy(ctx, permission.KindUser, "ops")
	assert.NoError(t, err)
	assert.Len(t, policy, 0)
	policy, err = policies.Policy(ctx, permission.KindUser, "dev")
	assert.NoError(t, err)
	assert.Len(t, policy, 1, "another subject keeps its entries")
}
