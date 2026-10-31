//go:build integration

package api_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/testkey"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestKey_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/keys"},
		{http.MethodPost, "/keys/generate"},
		{http.MethodPost, "/keys/upload"},
		{http.MethodDelete, "/keys/task_ssh_ed25519"},
	} {
		res := f.call(t, asOps, c.method, c.path, map[string]any{})
		assert.Equal(t, http.StatusForbidden, res.Status, c.method+" "+c.path)
	}
}

// Generate names a key by its algorithm and size; the listing shows
// every key, the instance's encryption key among them, never a private
// half. A name is unique across the folder.
func TestKey_GenerateAndList(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	list := f.call(t, asAdmin, http.MethodGet, "/keys", nil)
	assert.Equal(t, http.StatusOK, list.Status)
	assert.Len(t, list.list("keys"), 1, "the instance's encryption key")
	assert.Equal(t, f.opts.Keys.Dir, list.str("dir"))
	assert.Equal(t, "task_encryption_ed25519", list.str("encrypt", "name"))

	res := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Name: "deploy", Algo: "ecdsa", Bits: 384, Comment: " deploy "})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, "deploy_ecdsa384", res.str("key", "name"))
	assert.Equal(t, "ssh", res.str("key", "purpose"))
	assert.Equal(t, "instance", res.str("key", "scope"))
	assert.Equal(t, "deploy", res.str("key", "comment"))
	assert.Contains(t, res.str("key", "public_key"), "ecdsa-sha2-nistp384")
	assert.FileExists(t, filepath.Join(f.opts.Keys.Dir, "deploy_ecdsa384.key"))

	list = f.call(t, asAdmin, http.MethodGet, "/keys", nil)
	assert.Len(t, list.list("keys"), 2)
	assert.NotContains(t, string(list.Raw), "PRIVATE KEY")

	taken := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Name: "deploy", Type: "ed25519"})
	assert.Equal(t, http.StatusConflict, taken.Status)
	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Name: "x", Type: "dsa"}).Status)
	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Name: "Bad Name"}).Status)
}

// Upload reads the algorithm and the size from the key; the same key a
// second time, under any name or purpose, answers 409.
func TestKey_Upload(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	other := &keystore.Local{Dir: t.TempDir(), Rows: storage.NewKeyInfoStorage(f.db), Settings: storage.NewSettingStorage(f.db)}
	made, err := other.Import(context.Background(), keystore.Spec{Name: "elsewhere"}, testkey.Of(t, 2048))
	assert.NoError(t, err)
	pem, err := os.ReadFile(filepath.Join(other.Dir, made.Name+".key"))
	assert.NoError(t, err)
	assert.NoError(t, other.Rows.Delete(context.Background(), made.Name))

	res := f.call(t, asAdmin, http.MethodPost, "/keys/upload", api.KeyRequest{Name: "deploy", Content: string(pem)})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, "deploy_rsa2048", res.str("key", "name"))
	assert.Equal(t, made.Fingerprint, res.str("key", "fingerprint"))
	assert.NotContains(t, string(res.Raw), "PRIVATE KEY")

	again := f.call(t, asAdmin, http.MethodPost, "/keys/upload", api.KeyRequest{Name: "again", Purpose: "encrypt", Content: string(pem)})
	assert.Equal(t, http.StatusConflict, again.Status, "a key has one purpose")
	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPost, "/keys/upload", api.KeyRequest{Name: "junk", Content: "nope"}).Status)
}

// Delete retires an SSH key; the encryption key stays, it opens the
// secrets.
func TestKey_Delete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/keys/task_ssh_ed25519", nil).Status)
	assert.NoFileExists(t, filepath.Join(f.opts.Keys.Dir, "task_ssh_ed25519.key"))
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodDelete, "/keys/task_ssh_ed25519", nil).Status)
	assert.Equal(t, http.StatusConflict, f.call(t, asAdmin, http.MethodDelete, "/keys/task_encryption_ed25519", nil).Status)
}

// Rotating the instance's encryption key writes the integrations again;
// the old material is the key's one backup, backup/<name>, which goes
// through DELETE /keys/<name>/backup.
func TestKey_Secrets(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "integration:beszel", Name: "token", Content: "hunter2"}))
	old := f.opts.Keys.EncryptKey(ctx, storage.ScopeInstance)

	res := f.call(t, asAdmin, http.MethodPost, "/keys/encrypt/rotate", nil)
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.NotContains(t, string(res.Raw), "PRIVATE KEY")
	assert.Equal(t, float64(1), res.get("rekeyed"))
	assert.Equal(t, old.Name, res.str("key", "name"))
	assert.NotEqual(t, old.Fingerprint, res.str("key", "fingerprint"))
	assert.Equal(t, keystore.BackupName(old.Name), res.str("key", "backup"))
	assert.FileExists(t, filepath.Join(f.opts.Keys.Dir, "backup", old.Name+".key"))
	assert.Equal(t, http.StatusConflict, f.call(t, asAdmin, http.MethodPost, "/keys/encrypt/rotate", nil).Status, "a key with a backup rotates no further")
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/keys/"+old.Name+"/backup", nil).Status)
	assert.NoFileExists(t, filepath.Join(f.opts.Keys.Dir, "backup", old.Name+".key"))
	got, err := f.opts.Secrets.Get(ctx, model.Secret{Scope: "integration:beszel", Name: "token"})
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", got.Content, "rekeyed onto the new material")

	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPost, "/keys/encrypt/rotate", nil).Status)
}

// A key's links say whose runs get it: set on generate, changed on their
// own, and the git key besides; an encryption key links nothing.
func TestKey_Links(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	res := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Name: "deploy", Policies: []string{model.RoleAdmin}, Users: []string{" claude "}})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, []any{"admin"}, res.get("key", "roles"))
	assert.Equal(t, []any{"claude"}, res.get("key", "users"))

	yes := true
	put := f.call(t, asAdmin, http.MethodPut, "/keys/deploy_ed25519", api.KeyRequest{Policies: []string{model.RoleRunner}, Git: &yes})
	assert.Equal(t, http.StatusOK, put.Status, string(put.Raw))
	assert.Equal(t, []any{"runner"}, put.get("policies"))
	assert.Equal(t, []any{"claude"}, put.get("users"), "left out, the users stay")
	assert.Equal(t, "deploy_ed25519", f.call(t, asAdmin, http.MethodGet, "/keys", nil).str("git"))

	bad := f.call(t, asAdmin, http.MethodPut, "/keys/deploy_ed25519", api.KeyRequest{Policies: []string{"root"}})
	assert.Equal(t, http.StatusBadRequest, bad.Status)
	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPut, "/keys/task_encryption_ed25519", api.KeyRequest{Policies: []string{model.RoleAdmin}}).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodPut, "/keys/missing", api.KeyRequest{Policies: []string{}}).Status)
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPut, "/keys/deploy_ed25519", api.KeyRequest{}).Status)
}
