package app_test

import (
	"context"
	"net/http"
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

// Starting puts the integrations, encrypted with the SSH key, on the
// instance's own encryption key, after a copy of the
// database to .pre-keysplit; a second start does nothing. With the SSH
// key gone the integrations still open, and the SSH key is no age
// recipient.
func TestNew_MigratesIntegrations(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	data := t.TempDir()
	db := filepath.Join(data, "task-ui.db")
	keys := &sshkey.Store{Dir: t.TempDir()}
	_, err := keys.Generate("ed25519", "instance")
	assert.NoError(t, err)
	legacy := secret.NewVault(keys)
	encrypted, err := legacy.Encrypt("s3cret-value")
	assert.NoError(t, err)
	seed, err := storage.OpenAndMigrate(ctx, db)
	assert.NoError(t, err)
	assert.NoError(t, storage.NewSecretStorage(seed).Create(ctx, &model.Secret{Scope: "integration:beszel", Name: "token", Content: encrypted}))
	assert.NoError(t, seed.Close())

	cfg := &app.Config{Root: t.TempDir(), DataDir: data, DB: db, SSHDir: keys.Dir, SessionTTL: time.Hour, AutoLogin: "dev"}
	a, err := app.New(ctx, cfg, "test")
	assert.NoError(t, err)
	instance := cfg.Keyring().Instance()
	info, err := os.Stat(instance.Dir)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
	info, err = os.Stat(filepath.Join(instance.Dir, "id_ed25519"))
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	row, err := storage.NewSecretStorage(a.DB).Get(ctx, model.Secret{Scope: "integration:beszel", Name: "token"})
	assert.NoError(t, err)
	plain, err := secret.NewVault(instance).Decrypt(row.Content)
	assert.NoError(t, err)
	assert.Equal(t, "s3cret-value", plain)
	_, err = legacy.Decrypt(row.Content)
	assert.Error(t, err, "the SSH key opens it no more")
	copied, err := os.Stat(db + ".pre-keysplit")
	assert.NoError(t, err)
	a.Close()

	for _, name := range []string{"id_ed25519", "id_ed25519.pub"} {
		assert.NoError(t, os.Remove(filepath.Join(keys.Dir, name)))
	}
	b, err := app.New(ctx, cfg, "test")
	assert.NoError(t, err)
	again, err := storage.NewSecretStorage(b.DB).Get(ctx, model.Secret{Scope: "integration:beszel", Name: "token"})
	assert.NoError(t, err)
	assert.Equal(t, row.Content, again.Content, "a second start writes nothing")
	kept, err := os.Stat(db + ".pre-keysplit")
	assert.NoError(t, err)
	assert.Equal(t, copied.ModTime(), kept.ModTime())
	status, body, _ := get(t, b.Handler(), "/settings/integrations")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, body, `Token <span class="badge ok">set</span>`, "the token opens")
	assert.NotContains(t, body, "s3cret-value")
	b.Close()
}

// A key both encryption and SSH found at start shows as a warning on
// the keys page and stays in place.
func TestNew_KeyOverlap(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	data, keys := t.TempDir(), t.TempDir()
	cfg := &app.Config{Root: t.TempDir(), DataDir: data, DB: filepath.Join(data, "task-ui.db"), SSHDir: keys, SessionTTL: time.Hour, AutoLogin: "dev"}
	a, err := app.New(ctx, cfg, "test")
	assert.NoError(t, err)
	a.Close()
	instance := cfg.Keyring().Instance().Dir
	for _, suffix := range []string{"", ".pub"} {
		content, err := os.ReadFile(filepath.Join(instance, "id_ed25519"+suffix))
		assert.NoError(t, err)
		assert.NoError(t, os.WriteFile(filepath.Join(keys, "id_ed25519"+suffix), content, 0o600))
	}

	b, err := app.New(ctx, cfg, "test")
	assert.NoError(t, err)
	defer b.Close()
	status, body, _ := get(t, b.Handler(), "/settings/keys")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, body, `class="alert bad"`)
	assert.Contains(t, body, filepath.Join(keys, "id_ed25519"))
	assert.Contains(t, body, filepath.Join(instance, "id_ed25519"))
	assert.FileExists(t, filepath.Join(keys, "id_ed25519"))
	status, body, _ = get(t, b.Handler(), "/api/keys")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, body, `"keys":[]`)
	assert.Contains(t, body, `"overlaps":[{`)
}
