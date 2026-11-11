//go:build integration

package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// A key without a grant goes to every run of its workspace; a grant to a
// role or a user narrows it to theirs. Grants are per workspace.
func TestWorkspaceKeyGrantStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	grants := storage.NewWorkspaceKeyGrantStorage(newDB(t))
	keys := []string{"id_ed25519_deploy", "id_ed25519_admin", "id_rsa_alice"}

	got, err := grants.For(ctx, "ws1", keys, model.RoleRunner, "bob")
	assert.NoError(t, err)
	assert.Equal(t, keys, got, "no grant is every run")

	assert.NoError(t, grants.Set(ctx, "ws1", "id_ed25519_admin", []string{model.RoleAdmin, model.RoleAdmin}, nil))
	assert.NoError(t, grants.Set(ctx, "ws1", "id_rsa_alice", nil, []string{"alice", "", "alice"}))
	assert.NoError(t, grants.Set(ctx, "ws2", "id_ed25519_deploy", []string{model.RoleViewer}, nil))

	got, _ = grants.For(ctx, "ws1", keys, model.RoleRunner, "bob")
	assert.Equal(t, []string{"id_ed25519_deploy"}, got)
	got, _ = grants.For(ctx, "ws1", keys, model.RoleAdmin, "alice")
	assert.Equal(t, keys, got)
	got, _ = grants.For(ctx, "ws2", keys, model.RoleRunner, "bob")
	assert.Equal(t, []string{"id_ed25519_admin", "id_rsa_alice"}, got, "another workspace's grants")

	roles, users, err := grants.All(ctx, "ws1")
	assert.NoError(t, err)
	assert.Equal(t, map[string][]string{"id_ed25519_admin": {model.RoleAdmin}}, roles)
	assert.Equal(t, map[string][]string{"id_rsa_alice": {"alice"}}, users)

	assert.Error(t, grants.Set(ctx, "ws1", "id_rsa_alice", []string{"root"}, nil))
	assert.Error(t, grants.Set(ctx, "", "id_rsa_alice", nil, nil))

	// Set replaces; none opens the key to every run again.
	assert.NoError(t, grants.Set(ctx, "ws1", "id_ed25519_admin", nil, nil))
	got, _ = grants.For(ctx, "ws1", keys, model.RoleViewer, "")
	assert.Equal(t, []string{"id_ed25519_deploy", "id_ed25519_admin"}, got)

	assert.NoError(t, grants.Delete(ctx, "ws1", "id_rsa_alice"))
	assert.NoError(t, grants.DeleteWorkspace(ctx, "ws2"))
	roles, users, _ = grants.All(ctx, "ws1")
	assert.Len(t, roles, 0)
	assert.Len(t, users, 0)
	roles, _, _ = grants.All(ctx, "ws2")
	assert.Len(t, roles, 0)
}

// Every method returns the driver's error.
func TestWorkspaceKeyGrantStorage_Closed(t *testing.T) {
	t.Parallel()
	grants := storage.NewWorkspaceKeyGrantStorage(closedDB(t))
	ctx := context.Background()

	assert.ErrorContains(t, grants.Set(ctx, "ws1", "deploy", []string{model.RoleAdmin}, nil), errClosed)
	_, _, err := grants.All(ctx, "ws1")
	assert.ErrorContains(t, err, errClosed)
	_, err = grants.For(ctx, "ws1", []string{"deploy"}, model.RoleAdmin, "alice")
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, grants.Delete(ctx, "ws1", "deploy"), errClosed)
	assert.ErrorContains(t, grants.DeleteWorkspace(ctx, "ws1"), errClosed)
}

// Set clears the key's grants and then writes one row per role and per
// user, and stops at the write that failed.
func TestWorkspaceKeyGrantStorage_Writes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	refuse(t, db, model.WorkspaceKeyGrantTable, "INSERT")
	grants := storage.NewWorkspaceKeyGrantStorage(db)

	assert.ErrorContains(t, grants.Set(ctx, "ws1", "deploy", []string{model.RoleAdmin}, nil), errRefused)
	assert.ErrorContains(t, grants.Set(ctx, "ws1", "deploy", nil, []string{"alice"}), errRefused)
}
