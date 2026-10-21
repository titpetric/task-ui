package api_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
)

func TestKey_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/keys"},
		{http.MethodPost, "/keys/generate"},
		{http.MethodPost, "/keys/upload"},
		{http.MethodDelete, "/keys/id_ed25519"},
	} {
		res := f.call(t, asOps, c.method, c.path, map[string]any{})
		assert.Equal(t, http.StatusForbidden, res.Status, c.method+" "+c.path)
	}
}

func TestKey_GenerateAndList(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	empty := f.call(t, asAdmin, http.MethodGet, "/keys", nil)
	assert.Equal(t, http.StatusOK, empty.Status)
	assert.Len(t, empty.list("keys"), 0)
	assert.Equal(t, f.opts.Keys.Dir, empty.str("dir"))

	res := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "ed25519", Comment: " deploy "})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, "id_ed25519", res.str("key", "name"))
	assert.Equal(t, "deploy", res.str("key", "comment"))
	assert.Contains(t, res.str("key", "public_key"), "ssh-ed25519")
	assert.Nil(t, res.get("rekeyed"), "an SSH key encrypts nothing")

	list := f.call(t, asAdmin, http.MethodGet, "/keys", nil)
	assert.Len(t, list.list("keys"), 1)
	assert.Equal(t, f.opts.WorkspaceKeys.InstanceKey().Fingerprint, list.str("encrypt", "fingerprint"))
	assert.NotContains(t, string(list.Raw), "PRIVATE KEY")
	assert.Len(t, list.list("overlaps"), 0)

	// Generating again keeps the old pair as a backup.
	again := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "ed25519"})
	assert.Equal(t, http.StatusCreated, again.Status)
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/keys", nil).list("keys"), 2)

	bad := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "dsa"})
	assert.Equal(t, http.StatusBadRequest, bad.Status)
}

func TestKey_Upload(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	other := &sshkey.Store{Dir: t.TempDir()}
	key, err := other.Generate("ed25519", "elsewhere")
	assert.NoError(t, err)
	private, err := os.ReadFile(filepath.Join(other.Dir, key.Name))
	assert.NoError(t, err)

	res := f.call(t, asAdmin, http.MethodPost, "/keys/upload", api.KeyRequest{Content: string(private), Comment: "uploaded"})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, key.Fingerprint, res.str("key", "fingerprint"))
	assert.Equal(t, "uploaded", res.str("key", "comment"))

	bad := f.call(t, asAdmin, http.MethodPost, "/keys/upload", api.KeyRequest{Content: "not a key"})
	assert.Equal(t, http.StatusBadRequest, bad.Status)

	// An encryption key is no SSH key: the instance's, by its private
	// or its public half, or a workspace's.
	_, err = f.opts.WorkspaceKeys.Ensure("01ws", "task-ui ws")
	assert.NoError(t, err)
	for _, path := range []string{
		filepath.Join(f.opts.WorkspaceKeys.Instance().Dir, "id_ed25519"),
		filepath.Join(f.opts.WorkspaceKeys.Instance().Dir, "id_ed25519.pub"),
		filepath.Join(f.opts.WorkspaceKeys.Encrypt("01ws").Dir, "id_ed25519"),
	} {
		data, err := os.ReadFile(path)
		assert.NoError(t, err)
		refused := f.call(t, asAdmin, http.MethodPost, "/keys/upload", api.KeyRequest{Content: string(data), Name: "copy"})
		assert.Equal(t, http.StatusConflict, refused.Status, path)
		assert.Contains(t, refused.str("error"), "never both")
	}
	assert.NoFileExists(t, filepath.Join(f.opts.Keys.Dir, "id_ed25519_copy"))
}

func TestKey_Delete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	created := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "ed25519"})
	assert.Equal(t, http.StatusCreated, created.Status)

	res := f.call(t, asAdmin, http.MethodDelete, "/keys/"+created.str("key", "name"), nil)
	assert.Equal(t, http.StatusNoContent, res.Status)
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/keys", nil).list("keys"), 0)

	missing := f.call(t, asAdmin, http.MethodDelete, "/keys/id_rsa", nil)
	assert.Equal(t, http.StatusNotFound, missing.Status)
}

// An SSH key encrypts nothing: generating and deleting one leaves the
// integrations' settings as they are. Rotating the instance's
// encryption key writes them again for the new one and keeps the old
// as a backup; with the backup gone they still open.
func TestKey_Secrets(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	f.withKey(t)
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "integration:beszel", Name: "token", Content: "hunter2"}))
	before, err := f.opts.Secrets.Secrets.Get(ctx, model.Secret{Scope: "integration:beszel", Name: "token"})
	assert.NoError(t, err)

	rotated := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "ed25519", Comment: "rotated"})
	assert.Equal(t, http.StatusCreated, rotated.Status, string(rotated.Raw))
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/keys/id_ed25519", nil).Status)
	after, err := f.opts.Secrets.Secrets.Get(ctx, model.Secret{Scope: "integration:beszel", Name: "token"})
	assert.NoError(t, err)
	assert.Equal(t, before.Content, after.Content)

	old := f.opts.WorkspaceKeys.InstanceKey()
	res := f.call(t, asAdmin, http.MethodPost, "/keys/encrypt/rotate", nil)
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.NotContains(t, string(res.Raw), "PRIVATE KEY")
	assert.Equal(t, float64(1), res.get("rekeyed"))
	assert.NotEqual(t, old.Fingerprint, res.str("key", "fingerprint"))
	assert.Equal(t, old.Comment, res.str("key", "comment"))
	keys := f.opts.WorkspaceKeys.Instance().List()
	assert.Len(t, keys, 2)
	assert.True(t, keys[1].Backup)
	assert.NoError(t, f.opts.WorkspaceKeys.Instance().Delete(keys[1].Name))
	got, err := f.opts.Secrets.Get(ctx, model.Secret{Scope: "integration:beszel", Name: "token"})
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", got.Content)

	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPost, "/keys/encrypt/rotate", nil).Status)
}

// A key carries the roles whose runs get it: set on generate, changed on
// its own, kept by a rotation that names none, and gone with the key.
func TestKey_Policies(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "ed25519", Comment: "deploy", Policies: []string{"runner", "viewer"}})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	list := f.call(t, asAdmin, http.MethodGet, "/keys", nil)
	assert.Equal(t, []any{"runner", "viewer"}, list.get("policies", "id_ed25519"))

	// A rotation sends no policies, and keeps them.
	res = f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "ed25519", Comment: "deploy"})
	assert.Equal(t, http.StatusCreated, res.Status)
	assert.Equal(t, []any{"runner", "viewer"}, f.call(t, asAdmin, http.MethodGet, "/keys", nil).get("policies", "id_ed25519"))

	updated := f.call(t, asAdmin, http.MethodPut, "/keys/id_ed25519", api.KeyRequest{Policies: []string{"admin"}})
	assert.Equal(t, http.StatusOK, updated.Status, string(updated.Raw))
	assert.Equal(t, []any{"admin"}, f.call(t, asAdmin, http.MethodGet, "/keys", nil).get("policies", "id_ed25519"))

	none := f.call(t, asAdmin, http.MethodPut, "/keys/id_ed25519", map[string]any{})
	assert.Equal(t, http.StatusOK, none.Status)
	assert.Nil(t, f.call(t, asAdmin, http.MethodGet, "/keys", nil).get("policies", "id_ed25519"))

	bad := f.call(t, asAdmin, http.MethodPut, "/keys/id_ed25519", api.KeyRequest{Policies: []string{"root"}})
	assert.Equal(t, http.StatusBadRequest, bad.Status)
	missing := f.call(t, asAdmin, http.MethodPut, "/keys/id_rsa", api.KeyRequest{Policies: []string{"runner"}})
	assert.Equal(t, http.StatusNotFound, missing.Status)
	forbidden := f.call(t, asOps, http.MethodPut, "/keys/id_ed25519", api.KeyRequest{Policies: []string{"runner"}})
	assert.Equal(t, http.StatusForbidden, forbidden.Status)

	// Upload takes them too; a deleted key loses its own.
	pem, err := os.ReadFile(filepath.Join(f.opts.Keys.Dir, "id_ed25519"))
	assert.NoError(t, err)
	assert.NoError(t, f.opts.Keys.Delete("id_ed25519"))
	up := f.call(t, asAdmin, http.MethodPost, "/keys/upload", api.KeyRequest{Content: string(pem), Policies: []string{"runner"}})
	assert.Equal(t, http.StatusCreated, up.Status, string(up.Raw))
	assert.Equal(t, []any{"runner"}, f.call(t, asAdmin, http.MethodGet, "/keys", nil).get("policies", "id_ed25519"))
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/keys/id_ed25519", nil).Status)
	assert.Nil(t, f.call(t, asAdmin, http.MethodGet, "/keys", nil).get("policies", "id_ed25519"))
}

// An admin makes a key for a user: under a name, beside the type's own
// key, given to the user by name. A rotation keeps both.
func TestKey_ForAUser(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	own := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "ed25519", Comment: "instance"})
	assert.Equal(t, http.StatusCreated, own.Status, string(own.Raw))
	res := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "ed25519", Name: "ops", Comment: "ops", Users: []string{" ops "}})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	assert.Equal(t, "id_ed25519_ops", res.str("key", "name"))
	assert.Equal(t, "ops", res.str("key", "label"))

	list := f.call(t, asAdmin, http.MethodGet, "/keys", nil)
	assert.Len(t, list.list("keys"), 2)
	assert.Equal(t, []any{"ops"}, list.get("users", "id_ed25519_ops"))
	assert.Nil(t, list.get("users", "id_ed25519"))

	rotated := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "ed25519", Name: "ops", Comment: "ops"})
	assert.Equal(t, http.StatusCreated, rotated.Status)
	assert.Equal(t, []any{"ops"}, f.call(t, asAdmin, http.MethodGet, "/keys", nil).get("users", "id_ed25519_ops"))

	updated := f.call(t, asAdmin, http.MethodPut, "/keys/id_ed25519_ops", api.KeyRequest{Users: []string{"ops", "viewer"}, Policies: []string{"runner"}})
	assert.Equal(t, http.StatusOK, updated.Status, string(updated.Raw))
	after := f.call(t, asAdmin, http.MethodGet, "/keys", nil)
	assert.Equal(t, []any{"ops", "viewer"}, after.get("users", "id_ed25519_ops"))
	assert.Equal(t, []any{"runner"}, after.get("policies", "id_ed25519_ops"))

	bad := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "ed25519", Name: "Not OK"})
	assert.Equal(t, http.StatusBadRequest, bad.Status)
}
