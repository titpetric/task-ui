//go:build integration

package secret_test

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/testdb"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/storage"
)

// newStore returns a store over a private database and a folder with
// one ed25519 key.
func newStore(t *testing.T) (*secret.Store, *storage.SecretStorage) {
	t.Helper()
	db := testdb.New(t)
	keys := newKeys(t)
	_, err := keys.Generate("ed25519", "task-ui")
	assert.NoError(t, err)
	rows := storage.NewSecretStorage(db)
	return secret.New(rows, secret.NewVault(secret.DirKeys{Store: keys})), rows
}

func TestNew(t *testing.T) {
	t.Parallel()
	s, rows := newStore(t)
	assert.Equal(t, rows, s.Secrets)
	assert.NotNil(t, s.Vault)
}

func TestStore_SetAndRead(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s, rows := newStore(t)

	token := &model.Secret{Scope: storage.SharedScope(model.Default), Name: "TOKEN", Content: "hunter2"}
	assert.NoError(t, s.Set(ctx, token))
	assert.True(t, model.ValidID(token.ID))
	assert.Equal(t, "hunter2", token.Content)
	assert.NoError(t, s.Set(ctx, &model.Secret{Scope: "infra", Name: "TOKEN", Content: "infra value"}))
	assert.NoError(t, s.Set(ctx, &model.Secret{Scope: "infra", Name: "REGION", Content: "eu"}))

	// At rest the store keeps the content armored; every way out decrypts.
	raw, err := rows.Get(ctx, model.Secret{Scope: storage.SharedScope(model.Default), Name: "TOKEN"})
	assert.NoError(t, err)
	assert.True(t, secret.Encrypted(raw.Content))
	assert.NotContains(t, raw.Content, "hunter2")

	got, err := s.Get(ctx, model.Secret{Scope: storage.SharedScope(model.Default), Name: "TOKEN"})
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", got.Content)

	list, err := s.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 3)
	for _, item := range list {
		assert.False(t, secret.Encrypted(item.Content), item.Name)
	}

	infra, err := s.For(ctx, "infra")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"TOKEN": "infra value", "REGION": "eu"}, infra)
	website, err := s.For(ctx, "website")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"TOKEN": "hunter2"}, website)
	// The rows that reach infra, by name: REGION and its own TOKEN.
	reaching, err := s.Reaching(ctx, "infra")
	assert.NoError(t, err)
	assert.Len(t, reaching, 2)
	assert.Equal(t, "REGION", reaching[0].Name)
	assert.Equal(t, "infra", reaching[1].Scope)

	count, err := s.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	// Replacing keeps the row, deleting removes it.
	assert.NoError(t, s.Set(ctx, &model.Secret{Scope: storage.SharedScope(model.Default), Name: "TOKEN", Content: "changed"}))
	got, _ = s.Get(ctx, model.Secret{Scope: storage.SharedScope(model.Default), Name: "TOKEN"})
	assert.Equal(t, token.ID, got.ID)
	assert.Equal(t, "changed", got.Content)
	assert.NoError(t, s.Delete(ctx, token.ID))
	_, err = s.Get(ctx, model.Secret{Scope: storage.SharedScope(model.Default), Name: "TOKEN"})
	assert.ErrorIs(t, err, storage.ErrNotFound)

	_, err = s.Get(ctx, model.Secret{Scope: "x", Name: "NOPE"})
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.Error(t, s.Set(ctx, &model.Secret{Scope: "infra", Name: "bad name", Content: "x"}))
}

// Rekey writes every secret for the current keys: after a rotation the
// backup can go, and before a key goes the rest still open everything.
func TestStore_Rekey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s, rows := newStore(t)
	keys := s.Vault.Keys.(secret.DirKeys).Store
	assert.NoError(t, s.Set(ctx, &model.Secret{Scope: storage.SharedScope(model.Default), Name: "A", Content: "one"}))
	assert.NoError(t, s.Set(ctx, &model.Secret{Scope: "infra", Name: "B", Content: "two"}))
	before, _ := rows.List(ctx)

	_, err := keys.Generate("ed25519", "rotated")
	assert.NoError(t, err)
	n, err := s.Rekey(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	after, _ := rows.List(ctx)
	assert.NotEqual(t, before[0].Content, after[0].Content)

	for _, key := range keys.List() {
		if key.Backup {
			assert.NoError(t, keys.Delete(key.Name))
		}
	}
	infra, err := s.For(ctx, "infra")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"A": "one", "B": "two"}, infra)

	// Excluding the only key cannot rekey; every secret survives.
	_, err = s.Rekey(ctx, "id_ed25519")
	assert.ErrorIs(t, err, secret.ErrNoKey)
	infra, err = s.For(ctx, "infra")
	assert.NoError(t, err)
	assert.Len(t, infra, 2)

	// A secret no key opens fails every read with its name.
	assert.NoError(t, keys.Delete("id_ed25519"))
	_, err = keys.Generate("ed25519", "stranger")
	assert.NoError(t, err)
	for _, key := range keys.List() {
		if key.Backup {
			assert.NoError(t, keys.Delete(key.Name))
		}
	}
	_, err = s.List(ctx)
	assert.ErrorContains(t, err, "secret A for workspace:internal")
	_, err = s.For(ctx, "infra")
	assert.Error(t, err)
	_, err = s.Rekey(ctx)
	assert.Error(t, err)
}
