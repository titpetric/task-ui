package secret_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/storage"
)

// Migrate moves the shared secrets to workspace:_ and writes every
// workspace secret for its workspace's key, copying first; it leaves
// the integrations and a secret nothing opens as they are, and a second
// run does nothing.
func TestStore_Migrate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSpaces(t)
	legacy := secret.NewVault(secret.DirKeys{Store: s.legacy})
	put := func(scope, name, content string) {
		t.Helper()
		encrypted, err := legacy.Encrypt(content)
		assert.NoError(t, err)
		assert.NoError(t, s.rows.Create(ctx, &model.Secret{Scope: scope, Name: name, Content: encrypted}))
	}
	// As the instance had them: every secret on its SSH key.
	put(storage.ScopeAll, "SHARED", "root")
	put("web", "WEB", "w")
	put("acme/api", "API", "p")
	put("integration:claude", "token", "t")
	assert.NoError(t, s.rows.Create(ctx, &model.Secret{Scope: "web", Name: "PLAIN", Content: "unencrypted"}))
	stranger := newKeys(t)
	_, err := stranger.Generate("ed25519", "stranger")
	assert.NoError(t, err)
	lost, err := secret.NewVault(secret.DirKeys{Store: stranger}).Encrypt("lost")
	assert.NoError(t, err)
	assert.NoError(t, s.rows.Create(ctx, &model.Secret{Scope: "web", Name: "LOST", Content: lost}))
	integration, _ := s.rows.Get(ctx, model.Secret{Scope: "integration:claude", Name: "token"})

	copies := 0
	before := func(context.Context) error { copies++; return nil }
	result, err := s.store.Migrate(ctx, before)
	assert.NoError(t, err)
	assert.Equal(t, secret.Migration{Migrated: 4, Current: 0, Failed: 1}, result)
	assert.Equal(t, 1, copies)

	_, err = s.rows.Get(ctx, model.Secret{Scope: storage.ScopeAll, Name: "SHARED"})
	assert.ErrorIs(t, err, storage.ErrNotFound, "the shared row moved")
	still, _ := s.rows.Get(ctx, model.Secret{Scope: "integration:claude", Name: "token"})
	assert.Equal(t, integration.Content, still.Content, "MigrateInstance moves the integrations")
	left, _ := s.rows.Get(ctx, model.Secret{Scope: "web", Name: "LOST"})
	assert.Equal(t, lost, left.Content, "a secret nothing opens stays unchanged")

	hide(t, s.legacy)
	assert.NoError(t, s.rows.Delete(ctx, left.ID))
	web, err := s.store.For(ctx, "web")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"SHARED": "root", "WEB": "w", "PLAIN": "unencrypted"}, web)
	api, err := s.store.For(ctx, "acme/api")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"API": "p"}, api)

	again, err := s.store.Migrate(ctx, before)
	assert.NoError(t, err)
	assert.Equal(t, secret.Migration{Current: 4}, again)
	assert.Equal(t, 1, copies, "nothing to do copies nothing")
}

// A shared secret whose name workspace:_ already has stays in its scope.
func TestStore_MigrateConflict(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSpaces(t)
	encrypted, err := secret.NewVault(secret.DirKeys{Store: s.legacy}).Encrypt("old")
	assert.NoError(t, err)
	assert.NoError(t, s.rows.Create(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "TOKEN", Content: encrypted}))
	assert.NoError(t, s.store.Set(ctx, &model.Secret{Scope: "workspace:_", Name: "TOKEN", Content: "new"}))

	result, err := s.store.Migrate(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, secret.Migration{Current: 1, Failed: 1}, result)
	_, err = s.rows.Get(ctx, model.Secret{Scope: storage.ScopeAll, Name: "TOKEN"})
	assert.NoError(t, err)
	got, err := s.store.Get(ctx, model.Secret{Scope: "workspace:_", Name: "TOKEN"})
	assert.NoError(t, err)
	assert.Equal(t, "new", got.Content)
}

// MigrateInstance writes the integrations for the instance's
// encryption key, copying first: after it the legacy SSH keys can go
// and the integrations still open. A workspace's secrets stay as they
// are, an integration nothing opens stays and counts, and a second
// run does nothing.
func TestStore_MigrateInstance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSpaces(t)
	legacy := secret.NewVault(secret.DirKeys{Store: s.legacy})
	for name, content := range map[string]string{"url": "https://beszel", "token": "t"} {
		encrypted, err := legacy.Encrypt(content)
		assert.NoError(t, err)
		assert.NoError(t, s.rows.Create(ctx, &model.Secret{Scope: "integration:beszel", Name: name, Content: encrypted}))
	}
	assert.NoError(t, s.store.Set(ctx, &model.Secret{Scope: "integration:claude", Name: "token", Content: "current"}))
	assert.NoError(t, s.rows.Create(ctx, &model.Secret{Scope: "integration:plain", Name: "token", Content: "unencrypted"}))
	assert.NoError(t, s.store.Set(ctx, &model.Secret{Scope: "web", Name: "WEB", Content: "w"}))
	web, _ := s.rows.Get(ctx, model.Secret{Scope: "web", Name: "WEB"})
	stranger := newKeys(t)
	_, err := stranger.Generate("ed25519", "stranger")
	assert.NoError(t, err)
	lost, err := secret.NewVault(secret.DirKeys{Store: stranger}).Encrypt("lost")
	assert.NoError(t, err)
	assert.NoError(t, s.rows.Create(ctx, &model.Secret{Scope: "integration:lost", Name: "token", Content: lost}))

	copies := 0
	before := func(context.Context) error { copies++; return nil }
	result, err := s.store.MigrateInstance(ctx, before)
	assert.NoError(t, err)
	assert.Equal(t, secret.Migration{Migrated: 3, Current: 1, Failed: 1}, result)
	assert.Equal(t, 1, copies)
	left, _ := s.rows.Get(ctx, model.Secret{Scope: "integration:lost", Name: "token"})
	assert.Equal(t, lost, left.Content, "a secret nothing opens stays unchanged")
	after, _ := s.rows.Get(ctx, model.Secret{Scope: "web", Name: "WEB"})
	assert.Equal(t, web.Content, after.Content, "a workspace's secret stays")

	hide(t, s.legacy)
	assert.NoError(t, s.rows.Delete(ctx, left.ID))
	values, err := s.store.Scope(ctx, "integration:beszel")
	assert.NoError(t, err)
	assert.Len(t, values, 2)
	for _, v := range values {
		assert.Equal(t, map[string]string{"url": "https://beszel", "token": "t"}[v.Name], v.Content)
	}
	plain, err := s.store.Get(ctx, model.Secret{Scope: "integration:plain", Name: "token"})
	assert.NoError(t, err)
	assert.Equal(t, "unencrypted", plain.Content)
	row, _ := s.rows.Get(ctx, model.Secret{Scope: "integration:plain", Name: "token"})
	assert.True(t, secret.Encrypted(row.Content), "a plain value is encrypted on the way")

	again, err := s.store.MigrateInstance(ctx, before)
	assert.NoError(t, err)
	assert.Equal(t, secret.Migration{Current: 4}, again)
	assert.Equal(t, 1, copies, "nothing to do copies nothing")
}

// Without legacy keys an integration the instance's key does not open
// stays unchanged.
func TestStore_MigrateInstanceNoLegacy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSpaces(t)
	encrypted, err := secret.NewVault(secret.DirKeys{Store: s.legacy}).Encrypt("t")
	assert.NoError(t, err)
	assert.NoError(t, s.rows.Create(ctx, &model.Secret{Scope: "integration:claude", Name: "token", Content: encrypted}))
	s.store.Legacy = nil
	result, err := s.store.MigrateInstance(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, secret.Migration{Failed: 1}, result)
}
