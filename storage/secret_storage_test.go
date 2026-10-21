package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestValidName(t *testing.T) {
	t.Parallel()
	assert.True(t, storage.ValidName("TOKEN"))
	assert.True(t, storage.ValidName("_x1"))
	assert.False(t, storage.ValidName("1BAD"))
	assert.False(t, storage.ValidName("a-b"))
	assert.False(t, storage.ValidName(""))
}

func TestSecretStorage_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	secrets := storage.NewSecretStorage(newDB(t))
	secret := &model.Secret{Scope: storage.ScopeAll, Name: "TOKEN", Content: "x"}
	assert.NoError(t, secrets.Create(ctx, secret))
	assert.True(t, model.ValidID(secret.ID))
	assert.Error(t, secrets.Create(ctx, &model.Secret{Name: "TOKEN"}), "no scope")
	assert.ErrorIs(t, secrets.Create(ctx, &model.Secret{Scope: "infra", Name: "1BAD"}), storage.ErrInvalidName)
	assert.Error(t, secrets.Create(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "TOKEN"}), "duplicate")
}

func TestSecretStorage_Get(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	secrets := storage.NewSecretStorage(newDB(t))
	assert.NoError(t, secrets.Create(ctx, &model.Secret{Scope: "infra", Name: "REGION", Content: "eu"}))
	got, err := secrets.Get(ctx, model.Secret{Scope: "infra", Name: "REGION"})
	assert.NoError(t, err)
	assert.Equal(t, "eu", got.Content)
	_, err = secrets.Get(ctx, model.Secret{Scope: "web", Name: "REGION"})
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestSecretStorage_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	secrets := storage.NewSecretStorage(newDB(t))
	assert.NoError(t, secrets.Create(ctx, &model.Secret{Scope: "infra", Name: "TOKEN", Content: "mine"}))
	assert.NoError(t, secrets.Create(ctx, &model.Secret{Scope: "infra", Name: "REGION", Content: "eu"}))
	assert.NoError(t, secrets.Create(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "TOKEN", Content: "shared"}))
	assert.NoError(t, secrets.Create(ctx, &model.Secret{Scope: "web", Name: "CDN", Content: "c"}))
	assert.NoError(t, secrets.Create(ctx, &model.Secret{Scope: storage.SharedScope("acme"), Name: "ACME", Content: "a"}))

	list, err := secrets.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 5)
	assert.Equal(t, storage.ScopeAll, list[0].Scope, "shared first")
	assert.Equal(t, "workspace:acme", list[1].Scope, "a workspace's shared ones too")
	assert.Equal(t, "REGION", list[2].Name, "then by folder and name")

	scoped, err := secrets.Scoped(ctx, storage.ScopeAll, "infra")
	assert.NoError(t, err)
	assert.Len(t, scoped, 3)
	assert.Equal(t, storage.ScopeAll, scoped[0].Scope, "in the order of the scopes")
	assert.Equal(t, "REGION", scoped[1].Name, "then by name")
}

func TestSecretStorage_Set(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	secrets := storage.NewSecretStorage(newDB(t))
	first := &model.Secret{Scope: "infra", Name: "TOKEN", Content: "one"}
	assert.NoError(t, secrets.Set(ctx, first))
	second := &model.Secret{Scope: "infra", Name: "TOKEN", Content: "two"}
	assert.NoError(t, secrets.Set(ctx, second))
	assert.Equal(t, first.ID, second.ID)
	got, err := secrets.Get(ctx, *first)
	assert.NoError(t, err)
	assert.Equal(t, "two", got.Content)
	assert.ErrorIs(t, secrets.Set(ctx, &model.Secret{Scope: "infra", Name: "bad name"}), storage.ErrInvalidName)
}

func TestSecretStorage_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	secrets := storage.NewSecretStorage(newDB(t))
	secret := &model.Secret{Scope: "infra", Name: "TOKEN", Content: "one"}
	assert.NoError(t, secrets.Create(ctx, secret))
	secret.Content = "two"
	assert.NoError(t, secrets.Update(ctx, secret))
	got, err := secrets.Get(ctx, *secret)
	assert.NoError(t, err)
	assert.Equal(t, "two", got.Content)
	secret.Name = "bad name"
	assert.ErrorIs(t, secrets.Update(ctx, secret), storage.ErrInvalidName)
}

func TestSecretStorage_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	secrets := storage.NewSecretStorage(newDB(t))
	secret := &model.Secret{Scope: "infra", Name: "TOKEN", Content: "one"}
	assert.NoError(t, secrets.Create(ctx, secret))
	assert.NoError(t, secrets.Delete(ctx, secret.ID))
	_, err := secrets.Get(ctx, *secret)
	assert.ErrorIs(t, err, storage.ErrNotFound)
}
