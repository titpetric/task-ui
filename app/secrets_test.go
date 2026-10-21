package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// Starting puts the shared secrets on the default workspace's key,
// after a copy of the database; a second start finds nothing to do and
// keeps the copy.
func TestNew_MigratesSecrets(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	data := t.TempDir()
	db := filepath.Join(data, "task-ui.db")
	keys := &sshkey.Store{Dir: t.TempDir()}
	_, err := keys.Generate("ed25519", "instance")
	assert.NoError(t, err)
	encrypted, err := secret.NewVault(keys).Encrypt("hunter2")
	assert.NoError(t, err)
	seed, err := storage.OpenAndMigrate(ctx, db)
	assert.NoError(t, err)
	assert.NoError(t, storage.NewSecretStorage(seed).Create(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "TOKEN", Content: encrypted}))
	assert.NoError(t, seed.Close())

	cfg := &app.Config{Root: t.TempDir(), DataDir: data, DB: db, SSHDir: keys.Dir, SessionTTL: time.Hour, AutoLogin: "dev"}
	a, err := app.New(ctx, cfg, "test")
	assert.NoError(t, err)
	rows := storage.NewSecretStorage(a.DB)
	row, err := rows.Get(ctx, model.Secret{Scope: "workspace:_", Name: "TOKEN"})
	assert.NoError(t, err)
	_, err = secret.NewVault(keys).Decrypt(row.Content)
	assert.Error(t, err, "on the workspace's key, not the instance key")
	copied, err := os.Stat(db + ".pre-workspace-keys")
	assert.NoError(t, err)
	a.Close()

	b, err := app.New(ctx, cfg, "test")
	assert.NoError(t, err)
	again, err := storage.NewSecretStorage(b.DB).Get(ctx, model.Secret{Scope: "workspace:_", Name: "TOKEN"})
	assert.NoError(t, err)
	assert.Equal(t, row.Content, again.Content, "a second start writes nothing")
	kept, err := os.Stat(db + ".pre-workspace-keys")
	assert.NoError(t, err)
	assert.Equal(t, copied.ModTime(), kept.ModTime())
	b.Close()
}
