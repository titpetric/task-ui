//go:build integration

package secret_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// spaces is a store over a private database with two workspaces, the
// default one, _, and acme, every secret on the instance's encryption
// key, beside the legacy SSH keys the migrations read.
type spaces struct {
	store    *secret.Store
	rows     *storage.SecretStorage
	instance *sshkey.Store
	legacy   *sshkey.Store
	ids      map[string]string
}

func newSpaces(t *testing.T) *spaces {
	t.Helper()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	legacy := newKeys(t)
	_, err = legacy.Generate("ed25519", "legacy")
	assert.NoError(t, err)

	workspaces := storage.NewWorkspaceStorage(db)
	assert.NoError(t, workspaces.EnsureDefault(ctx))
	assert.NoError(t, workspaces.Create(ctx, &model.Workspace{Folder: "acme"}))
	instance := newKeys(t)
	_, err = instance.Generate("ed25519", "instance")
	assert.NoError(t, err)
	ids := map[string]string{}
	list, err := workspaces.List(ctx)
	assert.NoError(t, err)
	for _, ws := range list {
		ids[ws.Folder] = ws.ID
	}

	rows := storage.NewSecretStorage(db)
	store := secret.New(rows, secret.NewVault(secret.DirKeys{Store: instance}))
	store.Workspaces = &secret.Workspaces{Rows: workspaces}
	store.Legacy = secret.NewVault(secret.DirKeys{Store: legacy})
	return &spaces{store: store, rows: rows, instance: instance, legacy: legacy, ids: ids}
}

// hide moves a store's ed25519 pair away until the test ends.
func hide(t *testing.T, keys *sshkey.Store) {
	t.Helper()
	away := t.TempDir()
	for _, name := range []string{"id_ed25519", "id_ed25519.pub"} {
		assert.NoError(t, os.Rename(filepath.Join(keys.Dir, name), filepath.Join(away, name)))
	}
}

func TestWorkspaces(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSpaces(t)
	w := s.store.Workspaces

	for scope, want := range map[string]string{
		"workspace:acme": "acme", "workspace:internal": "internal", "workspace:_": "internal", "all": "internal",
		"acme": "acme", "acme/api": "acme", "web": "internal", "internal/web": "internal", "_/web": "internal", "acmes/x": "internal",
	} {
		got, err := w.Name(ctx, scope)
		assert.NoError(t, err)
		assert.Equal(t, want, got, scope)
	}
	row, err := w.Of(ctx, "acme/api")
	assert.NoError(t, err)
	assert.Equal(t, s.ids["acme"], row.ID)
	_, err = w.Of(ctx, "workspace:gone")
	assert.ErrorContains(t, err, "gone")

	assert.True(t, secret.Service("integration:claude"))
	assert.False(t, secret.Service("workspace:acme"))
	assert.False(t, secret.Service("acme/api"))

	var none *secret.Workspaces
	name, err := none.Name(ctx, "acme/api")
	assert.NoError(t, err)
	assert.Equal(t, "internal", name, "with no workspaces every folder is the default one's")
}

// Every secret, a workspace's and an integration's, opens with the
// instance's encryption key and nothing else. A folder gets its
// workspace's shared secrets and no other workspace's.
func TestStore_Workspaces(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSpaces(t)
	set := func(scope, name, content string) {
		t.Helper()
		assert.NoError(t, s.store.Set(ctx, &model.Secret{Scope: scope, Name: name, Content: content}))
	}
	set(storage.ScopeAll, "SHARED", "root")
	set("web", "WEB", "w")
	set("workspace:acme", "ACME", "a")
	set("acme/api", "API", "p")
	set("acme/api", "ACME", "mine")
	set("integration:claude", "token", "t")

	shared, err := s.rows.Get(ctx, model.Secret{Scope: "workspace:internal", Name: "SHARED"})
	assert.NoError(t, err, "all is the default workspace's shared scope")
	plain, err := secret.NewVault(secret.DirKeys{Store: s.instance}).Decrypt(shared.Content)
	assert.NoError(t, err, "the instance key opens a workspace's secret")
	assert.Equal(t, "root", plain)
	_, err = secret.NewVault(secret.DirKeys{Store: s.legacy}).Decrypt(shared.Content)
	assert.Error(t, err, "the legacy keys open nothing new")

	web, err := s.store.For(ctx, "web")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"SHARED": "root", "WEB": "w"}, web)
	api, err := s.store.For(ctx, "acme/api")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"ACME": "mine", "API": "p"}, api, "no _ secrets in acme, the folder's own wins")

	n, err := s.store.OnInstance(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 6, n, "every secret is on the instance key")

	// With the instance key away nothing opens.
	hide(t, s.instance)
	_, err = s.store.For(ctx, "web")
	assert.Error(t, err)
	_, err = s.store.Scope(ctx, "integration:claude")
	assert.Error(t, err)
	_, err = s.store.Get(ctx, model.Secret{Scope: "acme/api", Name: "API"})
	assert.Error(t, err)
}

// Rekey writes every secret for the instance's key; Forget removes a
// workspace's secrets and no other.
func TestStore_Forget(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSpaces(t)
	assert.NoError(t, s.store.Set(ctx, &model.Secret{Scope: "workspace:acme", Name: "A", Content: "one"}))
	assert.NoError(t, s.store.Set(ctx, &model.Secret{Scope: "acme/api", Name: "B", Content: "two"}))
	assert.NoError(t, s.store.Set(ctx, &model.Secret{Scope: "web", Name: "C", Content: "three"}))
	api, err := s.store.For(ctx, "acme/api")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"A": "one", "B": "two"}, api)

	n, err := s.store.Rekey(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 3, n, "Rekey writes every scope with its own key")

	n, err = s.store.Forget(ctx, "acme")
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	count, _ := s.store.Count(ctx)
	assert.Equal(t, 1, count)
}

// RekeyInstance writes every secret, the integrations' settings and
// the workspaces' secrets, for a rotated instance key; Workspace names
// a scope's workspace.
func TestStore_RekeyInstance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSpaces(t)
	assert.NoError(t, s.store.Set(ctx, &model.Secret{Scope: "integration:claude", Name: "token", Content: "t"}))
	assert.NoError(t, s.store.Set(ctx, &model.Secret{Scope: "web", Name: "WEB", Content: "w"}))
	web, _ := s.rows.Get(ctx, model.Secret{Scope: "web", Name: "WEB"})

	_, err := s.instance.Generate("ed25519", "rotated")
	assert.NoError(t, err)
	n, err := s.store.RekeyInstance(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	after, _ := s.rows.Get(ctx, model.Secret{Scope: "web", Name: "WEB"})
	assert.NotEqual(t, web.Content, after.Content, "written for the new material")
	values, err := s.store.Scope(ctx, "integration:claude")
	assert.NoError(t, err)
	assert.Equal(t, "t", values[0].Content)
	got, err := s.store.Get(ctx, model.Secret{Scope: "web", Name: "WEB"})
	assert.NoError(t, err)
	assert.Equal(t, "w", got.Content)

	name, err := s.store.Workspace(ctx, "acme/api")
	assert.NoError(t, err)
	assert.Equal(t, "acme", name)
}

// Plain returns a workspace's rows decrypted, for an export, none of
// another workspace's or a service's.
func TestStore_Plain(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSpaces(t)
	for _, row := range []model.Secret{
		{Scope: "workspace:acme", Name: "SHARED", Content: "s"},
		{Scope: "acme/api", Name: "OWN", Content: "o"},
		{Scope: "web", Name: "ROOT", Content: "r"},
		{Scope: "integration:beszel", Name: "TOKEN", Content: "t"},
	} {
		assert.NoError(t, s.store.Set(ctx, &row))
	}
	rows, err := s.store.Plain(ctx, "acme")
	assert.NoError(t, err)
	plain := map[string]string{}
	for _, row := range rows {
		plain[row.Name] = row.Content
	}
	assert.Equal(t, map[string]string{"SHARED": "s", "OWN": "o"}, plain)
}

// Unopened counts a workspace's rows the instance key does not open:
// plain text, or encrypted to another key.
func TestStore_Unopened(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSpaces(t)
	assert.NoError(t, s.store.Set(ctx, &model.Secret{Scope: "acme/api", Name: "OWN", Content: "o"}))
	n, err := s.store.Unopened(ctx, "acme")
	assert.NoError(t, err)
	assert.Equal(t, 0, n)

	legacy, err := s.store.Legacy.Encrypt("old")
	assert.NoError(t, err)
	assert.NoError(t, s.rows.Set(ctx, &model.Secret{Scope: "workspace:acme", Name: "OLD", Content: legacy}))
	assert.NoError(t, s.rows.Set(ctx, &model.Secret{Scope: "acme/api", Name: "RAW", Content: "plain"}))
	assert.NoError(t, s.rows.Set(ctx, &model.Secret{Scope: "web", Name: "ELSEWHERE", Content: "plain"}))
	n, err = s.store.Unopened(ctx, "acme")
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
}

// Encrypted returns a workspace's rows as stored, its shared ones and
// its folders', none of another workspace's or a service's.
func TestStore_Encrypted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSpaces(t)
	for _, row := range []model.Secret{
		{Scope: "workspace:acme", Name: "SHARED", Content: "s"},
		{Scope: "acme/api", Name: "OWN", Content: "o"},
		{Scope: "web", Name: "ROOT", Content: "r"},
		{Scope: "integration:beszel", Name: "TOKEN", Content: "t"},
	} {
		assert.NoError(t, s.store.Set(ctx, &row))
	}
	rows, err := s.store.Encrypted(ctx, "acme")
	assert.NoError(t, err)
	assert.Len(t, rows, 2)
	for _, row := range rows {
		assert.True(t, secret.Encrypted(row.Content), row.Name)
		assert.True(t, row.Scope == "workspace:acme" || row.Scope == "acme/api", row.Scope)
	}
}
