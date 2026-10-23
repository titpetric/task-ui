package keystore_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/storage"
)

// A workspace's encryption key goes for good: its row, its files and
// its backups. The instance's stays.
func TestLocal_Purge(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := newLocal(t)
	instance, err := l.Generate(ctx, keystore.Spec{Name: "task_encryption", Purpose: keystore.PurposeEncrypt})
	assert.NoError(t, err)
	ws, err := l.Generate(ctx, keystore.Spec{Name: "workspace_acme_encryption", Purpose: keystore.PurposeEncrypt, Scope: keystore.ScopeWorkspace("w1")})
	assert.NoError(t, err)
	_, err = l.Rotate(ctx, ws.Name)
	assert.NoError(t, err)
	older := filepath.Join(l.Dir, ws.Name+".key.bak-20260101T000000")
	assert.NoError(t, os.WriteFile(older, []byte("old"), 0o600))

	keys, err := l.WorkspaceEncryptKeys(ctx)
	assert.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, ws.Name, keys[0].Name)

	assert.NoError(t, l.Purge(ctx, ws.Name))
	for _, path := range []string{
		filepath.Join(l.Dir, ws.Name+".key"), filepath.Join(l.Dir, ws.Name+".pub"),
		filepath.Join(l.Dir, "backup", ws.Name+".key"), filepath.Join(l.Dir, "backup", ws.Name+".pub"), older,
	} {
		_, err := os.Stat(path)
		assert.True(t, os.IsNotExist(err), path)
	}
	_, err = l.Get(ctx, ws.Name)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.ErrorIs(t, l.Purge(ctx, ws.Name), storage.ErrNotFound)
	keys, err = l.WorkspaceEncryptKeys(ctx)
	assert.NoError(t, err)
	assert.Len(t, keys, 0)
	assert.NotNil(t, l.EncryptKey(ctx, storage.ScopeInstance), "the instance's key stays")
	_, err = os.Stat(filepath.Join(l.Dir, instance.Name+".key"))
	assert.NoError(t, err)
}

// An export bundles a workspace's SSH keys alone; an older bundle's
// encryption key opens its secrets and stays out of the key folder.
func TestLocal_BundleSSHOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := newLocal(t)
	scope := keystore.ScopeWorkspace("w1")
	_, err := l.Generate(ctx, keystore.Spec{Name: "workspace_acme_encryption", Purpose: keystore.PurposeEncrypt, Scope: scope})
	assert.NoError(t, err)
	_, err = l.Generate(ctx, keystore.Spec{Name: "workspace_acme_ssh", Scope: scope})
	assert.NoError(t, err)
	files, _, _, err := l.Bundle(ctx, scope)
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	_, ok := files["ssh/workspace_acme_ssh_ed25519.key"]
	assert.True(t, ok)

	old, err := l.Private(ctx, "workspace_acme_encryption_ed25519")
	assert.NoError(t, err)
	files["encrypt/workspace_acme_encryption_ed25519.key"] = old
	assert.Len(t, keystore.BundledEncryptKeys(files), 1)

	other := newLocal(t)
	assert.NoError(t, other.CheckBundle(ctx, files))
	taken, err := other.Restore(ctx, keystore.ScopeWorkspace("w2"), "acme", "beta", files, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"workspace_beta_ssh_ed25519"}, taken)
	keys, err := other.WorkspaceEncryptKeys(ctx)
	assert.NoError(t, err)
	assert.Len(t, keys, 0, "no workspace encryption key comes in")
}
