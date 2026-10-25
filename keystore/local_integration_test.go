//go:build integration

package keystore_test

import (
	"context"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/storage"
)

// newStore returns a key folder over an in-memory database.
func newStore(t *testing.T) *keystore.Local {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return &keystore.Local{Dir: filepath.Join(t.TempDir(), "keys"), Rows: storage.NewKeyInfoStorage(db), Settings: storage.NewSettingStorage(db), Tmp: t.TempDir()}
}

// Import reads the algorithm and the size from the key, and refuses a
// key held under any name or purpose: a key has one purpose.
func TestLocal_Import(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := newStore(t)
	other := newStore(t)
	made, err := other.Generate(ctx, keystore.Spec{Name: "deploy", Algo: "rsa", Bits: 2048})
	assert.NoError(t, err)
	pem, err := other.Private(ctx, made.Name)
	assert.NoError(t, err)

	key, err := l.Import(ctx, keystore.Spec{Name: "deploy"}, pem)
	assert.NoError(t, err)
	assert.Equal(t, "deploy_rsa2048", key.Name)
	assert.Equal(t, "rsa", key.Algo)
	assert.Equal(t, 2048, key.Bits)
	assert.Equal(t, made.Fingerprint, key.Fingerprint)

	_, err = l.Import(ctx, keystore.Spec{Name: "again", Purpose: keystore.PurposeEncrypt}, pem)
	assert.ErrorIs(t, err, keystore.ErrExists, "the same key as an encryption key")
	_, err = l.Import(ctx, keystore.Spec{Name: "junk"}, []byte("not a key"))
	assert.Error(t, err)
}

// Rotate writes new material for a key and keeps the old as its one
// backup, backup/<name>.key, which the key names;
// Remove retires the files, drops the row, and the git setting with it.
func TestLocal_RotateRemoveGit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := newStore(t)
	key, err := l.Generate(ctx, keystore.Spec{Name: "task_ssh"})
	assert.NoError(t, err)
	assert.Nil(t, l.Git(ctx))
	assert.NoError(t, l.SetGit(ctx, key.Name))
	assert.Equal(t, key.Name, l.Git(ctx).Name)
	a, public, err := l.GitAgent(ctx)
	assert.NoError(t, err)
	assert.Equal(t, l.PublicPath(key.Name), public)
	assert.Equal(t, filepath.Join(l.Dir, key.Name+".pub"), public)
	assert.NoError(t, a.Close())
	assert.Equal(t, filepath.Join(l.Dir, "known_hosts"), l.KnownHosts())

	rotated, err := l.Rotate(ctx, key.Name)
	assert.NoError(t, err)
	assert.Equal(t, key.Name, rotated.Name)
	assert.NotEqual(t, key.Fingerprint, rotated.Fingerprint)
	assert.FileExists(t, filepath.Join(l.Dir, "backup", key.Name+".key"))
	assert.Equal(t, keystore.BackupName(key.Name), rotated.Backup)
	assert.True(t, l.Git(ctx).Git)

	enc, err := l.Generate(ctx, keystore.Spec{Name: "task_encryption", Purpose: keystore.PurposeEncrypt})
	assert.NoError(t, err)
	assert.ErrorIs(t, l.SetGit(ctx, enc.Name), keystore.ErrNotSSH)

	assert.NoError(t, l.Remove(ctx, key.Name))
	assert.Nil(t, l.Git(ctx))
	_, err = l.Get(ctx, key.Name)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.NoFileExists(t, filepath.Join(l.Dir, key.Name+".key"))
	assert.ErrorIs(t, l.Remove(ctx, key.Name), storage.ErrNotFound)

	ws, err := l.Generate(ctx, keystore.Spec{Name: "w", Scope: keystore.ScopeWorkspace("w1")})
	assert.NoError(t, err)
	assert.Equal(t, "w1", ws.Workspace())
	assert.NoError(t, l.RemoveScope(ctx, keystore.ScopeWorkspace("w1")))
	keys, _ := l.List(ctx)
	assert.Len(t, keys, 1)
}

// The vault keys of a scope are its encryption key and its backups.
func TestLocal_Vault(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := newStore(t)
	v := l.Vault(storage.ScopeInstance)
	_, _, ok := v.Recipient()
	assert.False(t, ok)
	assert.Len(t, v.Identities(), 0)
	key, err := l.Generate(ctx, keystore.Spec{Name: "task_encryption", Purpose: keystore.PurposeEncrypt})
	assert.NoError(t, err)
	name, public, ok := v.Recipient()
	assert.True(t, ok)
	assert.Equal(t, key.Name, name)
	assert.Equal(t, key.PublicKey, public)
	_, err = l.Rotate(ctx, key.Name)
	assert.NoError(t, err)
	assert.Len(t, v.Identities(), 2, "the backup opens what the old material encrypted")
}
