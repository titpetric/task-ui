package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// A start makes the key folder private, takes the keys of the older
// layout into it once, named, and makes the instance's encryption key
// when there is none; the old files stay.
func TestNew_Keys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	data, keys := t.TempDir(), t.TempDir()
	legacy := &sshkey.Store{Dir: filepath.Join(data, "instance", "encrypt")}
	old, err := legacy.Generate("ed25519", "task-ui instance")
	assert.NoError(t, err)
	own, err := (&sshkey.Store{Dir: keys}).Generate("ed25519", "task@task")
	assert.NoError(t, err)
	cfg := &app.Config{Root: t.TempDir(), DataDir: data, DB: filepath.Join(data, "task-ui.db"), SSHDir: keys, SessionTTL: time.Hour, AutoLogin: "dev"}
	a, err := app.New(ctx, cfg, "test")
	assert.NoError(t, err)
	defer a.Close()

	info, err := os.Stat(keys)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
	folder := &keystore.Local{Dir: keys, Rows: storage.NewKeyInfoStorage(a.DB), Settings: storage.NewSettingStorage(a.DB)}
	instance := folder.EncryptKey(ctx, storage.ScopeInstance)
	assert.Equal(t, "task_encryption_ed25519", instance.Name)
	assert.Equal(t, old.Fingerprint, instance.Fingerprint, "the old key, not a new one")
	git := folder.Git(ctx)
	assert.Equal(t, "task_ssh_ed25519", git.Name)
	assert.Equal(t, own.Fingerprint, git.Fingerprint)
	assert.FileExists(t, filepath.Join(keys, "id_ed25519"), "the old file stays")
}
