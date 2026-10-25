package keystore_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/storage"
)

// Rotate key keeps the previous material as the key's one backup, which
// serves beside the new key until it goes; a second rotation waits for
// that.
func TestLocal_Rotate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := newLocal(t)
	key, err := l.Generate(ctx, keystore.Spec{Name: "deploy", Algo: "rsa", Bits: 2048})
	assert.NoError(t, err)
	old := key.Fingerprint
	assert.NoError(t, l.SetGit(ctx, key.Name))

	rotated, err := l.Rotate(ctx, key.Name)
	assert.NoError(t, err)
	assert.Equal(t, key.Name, rotated.Name)
	assert.Equal(t, "rsa", rotated.Algo)
	assert.Equal(t, 2048, rotated.Bits)
	assert.NotEqual(t, old, rotated.Fingerprint)
	assert.Equal(t, keystore.BackupName(key.Name), rotated.Backup)
	assert.Equal(t, "backup/deploy_rsa2048", rotated.Backup)
	assert.Equal(t, old, rotated.BackupFingerprint)
	assert.NotNil(t, rotated.RotatedAt)
	assert.True(t, rotated.Git, "the git key stays the git key")

	dir := filepath.Join(l.Dir, "backup")
	info, err := os.Stat(dir)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
	for _, ext := range []string{".key", ".pub"} {
		info, err := os.Stat(filepath.Join(dir, "deploy_rsa2048"+ext))
		assert.NoError(t, err, ext)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), ext)
	}

	agent, err := l.AgentKeys(ctx, []string{key.Name})
	assert.NoError(t, err)
	assert.Len(t, agent, 2, "the agent serves the key and its backup")
	assert.Equal(t, key.Name, agent[0].Name, "the new key first")
	assert.Equal(t, "backup/deploy_rsa2048", agent[1].Name)

	_, err = l.Rotate(ctx, key.Name)
	assert.ErrorIs(t, err, keystore.ErrBackup, "one backup a key")

	assert.NoError(t, l.DeleteBackup(ctx, key.Name))
	_, err = os.Stat(filepath.Join(dir, "deploy_rsa2048.key"))
	assert.True(t, os.IsNotExist(err))
	assert.ErrorIs(t, l.DeleteBackup(ctx, key.Name), keystore.ErrNoBackup)
	agent, err = l.AgentKeys(ctx, []string{key.Name})
	assert.NoError(t, err)
	assert.Len(t, agent, 1)

	again, err := l.Rotate(ctx, key.Name)
	assert.NoError(t, err, "rotate works once the backup is gone")
	assert.Equal(t, rotated.Fingerprint, again.BackupFingerprint)

	data, err := json.Marshal(again)
	assert.NoError(t, err)
	assert.False(t, strings.Contains(string(data), "PRIVATE KEY"), "no private material in a key's view")
}

// The encryption key decrypts with its backup while it has one, so a
// secret written before the rotation still opens.
func TestLocal_RotateEncrypt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	l := newLocal(t)
	key, err := l.Generate(ctx, keystore.Spec{Name: "task_encryption", Purpose: keystore.PurposeEncrypt})
	assert.NoError(t, err)
	vault := secret.NewVault(l.Vault(storage.ScopeInstance))
	text, err := vault.Encrypt("hunter2")
	assert.NoError(t, err)

	_, err = l.Rotate(ctx, key.Name)
	assert.NoError(t, err)
	plain, err := vault.Decrypt(text)
	assert.NoError(t, err, "the backup decrypts what the previous key encrypted")
	assert.Equal(t, "hunter2", plain)

	_, err = l.AgentKeys(ctx, []string{key.Name})
	assert.ErrorIs(t, err, keystore.ErrNotSSH)
}

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
