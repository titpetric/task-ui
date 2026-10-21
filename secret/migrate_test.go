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
	instance := secret.NewVault(s.instance)
	put := func(scope, name, content string) {
		t.Helper()
		encrypted, err := instance.Encrypt(content)
		assert.NoError(t, err)
		assert.NoError(t, s.rows.Create(ctx, &model.Secret{Scope: scope, Name: name, Content: encrypted}))
	}
	// As the instance had them: every secret on the instance key.
	put(storage.ScopeAll, "SHARED", "root")
	put("web", "WEB", "w")
	put("acme/api", "API", "p")
	put("integration:claude", "token", "t")
	assert.NoError(t, s.rows.Create(ctx, &model.Secret{Scope: "web", Name: "PLAIN", Content: "unencrypted"}))
	stranger := newKeys(t)
	_, err := stranger.Generate("ed25519", "stranger")
	assert.NoError(t, err)
	lost, err := secret.NewVault(stranger).Encrypt("lost")
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
	assert.Equal(t, integration.Content, still.Content, "an integration stays on the instance key")
	left, _ := s.rows.Get(ctx, model.Secret{Scope: "web", Name: "LOST"})
	assert.Equal(t, lost, left.Content, "a secret nothing opens stays unchanged")

	s.hideInstance(t)
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
	encrypted, err := secret.NewVault(s.instance).Encrypt("old")
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
