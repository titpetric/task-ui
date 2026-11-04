//go:build integration

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

// Every method hands back what the driver said.
func TestSSHKeyPolicyStorage_Closed(t *testing.T) {
	t.Parallel()
	policies := storage.NewSSHKeyPolicyStorage(closedDB(t))
	ctx := context.Background()

	assert.ErrorContains(t, policies.Set(ctx, "deploy", []string{model.RoleAdmin}), errClosed)
	assert.ErrorContains(t, policies.SetUsers(ctx, "deploy", []string{"alice"}), errClosed)
	_, err := policies.Roles(ctx, "deploy")
	assert.ErrorContains(t, err, errClosed)
	_, err = policies.All(ctx)
	assert.ErrorContains(t, err, errClosed)
	_, err = policies.Keys(ctx, model.RoleAdmin)
	assert.ErrorContains(t, err, errClosed)
	_, err = policies.AllUsers(ctx)
	assert.ErrorContains(t, err, errClosed)
	_, err = policies.For(ctx, model.RoleAdmin, "alice")
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, policies.Delete(ctx, "deploy"), errClosed)
}

// Set and SetUsers clear the key's rows and then write them, and stop at
// the write that failed. Delete clears two tables and stops at the
// second: the trigger leaves the users alone and refuses the roles.
func TestSSHKeyPolicyStorage_Writes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	roles := newDB(t)
	refuse(t, roles, model.SSHKeyPolicyTable, "INSERT")
	assert.ErrorContains(t, storage.NewSSHKeyPolicyStorage(roles).Set(ctx, "deploy", []string{model.RoleAdmin}), errRefused)

	users := newDB(t)
	refuse(t, users, model.SSHKeyUserTable, "INSERT")
	assert.ErrorContains(t, storage.NewSSHKeyPolicyStorage(users).SetUsers(ctx, "deploy", []string{"alice"}), errRefused)

	second := newDB(t)
	held := storage.NewSSHKeyPolicyStorage(second)
	assert.NoError(t, held.Set(ctx, "deploy", []string{model.RoleAdmin}))
	refuse(t, second, model.SSHKeyPolicyTable, "DELETE")
	assert.ErrorContains(t, held.Delete(ctx, "deploy"), errRefused)
}
