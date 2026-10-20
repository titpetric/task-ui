package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Roles get a key by its name; a key can be in many roles' sets, and a
// role can have many keys.
func TestSSHKeyPolicyStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	policies := storage.NewSSHKeyPolicyStorage(newDB(t))

	assert.NoError(t, policies.Set(ctx, "id_ed25519", []string{model.RoleViewer, model.RoleRunner, model.RoleRunner}))
	assert.NoError(t, policies.Set(ctx, "id_rsa", []string{model.RoleRunner}))

	roles, err := policies.Roles(ctx, "id_ed25519")
	assert.NoError(t, err)
	assert.Equal(t, []string{model.RoleRunner, model.RoleViewer}, roles) // most permissive first, once

	keys, err := policies.Keys(ctx, model.RoleRunner)
	assert.NoError(t, err)
	assert.Equal(t, []string{"id_ed25519", "id_rsa"}, keys)
	keys, _ = policies.Keys(ctx, model.RoleAdmin)
	assert.Len(t, keys, 0)

	all, err := policies.All(ctx)
	assert.NoError(t, err)
	assert.Equal(t, map[string][]string{"id_ed25519": {model.RoleRunner, model.RoleViewer}, "id_rsa": {model.RoleRunner}}, all)

	// Set replaces; none leaves the key with nobody.
	assert.NoError(t, policies.Set(ctx, "id_ed25519", []string{model.RoleAdmin}))
	roles, _ = policies.Roles(ctx, "id_ed25519")
	assert.Equal(t, []string{model.RoleAdmin}, roles)
	assert.NoError(t, policies.Set(ctx, "id_rsa", nil))
	roles, _ = policies.Roles(ctx, "id_rsa")
	assert.Len(t, roles, 0)

	assert.Error(t, policies.Set(ctx, "id_rsa", []string{"root"}))
	assert.Error(t, policies.Set(ctx, "", []string{model.RoleRunner}))

	assert.NoError(t, policies.Delete(ctx, "id_ed25519"))
	keys, _ = policies.Keys(ctx, model.RoleAdmin)
	assert.Len(t, keys, 0)
}

// A key given to users by name reaches them beside their role's keys.
func TestSSHKeyPolicyStorage_Users(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	policies := storage.NewSSHKeyPolicyStorage(newDB(t))
	assert.NoError(t, policies.Set(ctx, "id_ed25519", []string{model.RoleRunner}))
	assert.NoError(t, policies.SetUsers(ctx, "id_ed25519_alice", []string{"alice", "alice", ""}))
	assert.NoError(t, policies.SetUsers(ctx, "id_rsa_ops", []string{"bob", "alice"}))

	keys, err := policies.For(ctx, model.RoleRunner, "alice")
	assert.NoError(t, err)
	assert.Equal(t, []string{"id_ed25519", "id_ed25519_alice", "id_rsa_ops"}, keys)
	keys, _ = policies.For(ctx, model.RoleViewer, "bob")
	assert.Equal(t, []string{"id_rsa_ops"}, keys)
	keys, _ = policies.For(ctx, model.RoleViewer, "carol")
	assert.Len(t, keys, 0)

	all, err := policies.AllUsers(ctx)
	assert.NoError(t, err)
	assert.Equal(t, map[string][]string{"id_ed25519_alice": {"alice"}, "id_rsa_ops": {"alice", "bob"}}, all)

	assert.NoError(t, policies.SetUsers(ctx, "id_rsa_ops", nil))
	keys, _ = policies.For(ctx, model.RoleViewer, "bob")
	assert.Len(t, keys, 0)
	assert.Error(t, policies.SetUsers(ctx, "", []string{"alice"}))

	assert.NoError(t, policies.Delete(ctx, "id_ed25519_alice"))
	keys, _ = policies.For(ctx, model.RoleViewer, "alice")
	assert.Len(t, keys, 0)
}
