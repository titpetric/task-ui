package api_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestWorkspaceKey_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	assert.NoError(t, f.opts.Workspaces.Create(context.Background(), &model.Workspace{Folder: "malta"}))

	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/workspaces/malta/keys"},
		{http.MethodPost, "/workspaces/malta/keys/generate"},
		{http.MethodPost, "/workspaces/malta/keys/upload"},
		{http.MethodPut, "/workspaces/malta/keys/id_ed25519"},
		{http.MethodDelete, "/workspaces/malta/keys/id_ed25519"},
	} {
		for _, user := range []string{asOps, asViewer, asLimited} {
			res := f.call(t, user, c.method, c.path, map[string]any{})
			assert.Equal(t, http.StatusForbidden, res.Status, user+" "+c.method+" "+c.path)
		}
	}
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/workspaces/nope/keys", nil).Status)
}

// A workspace made gets its encryption key; an admin generates,
// uploads, grants and deletes its SSH keys, and no answer carries a
// private key.
// Deleting the workspace takes its keys.
func TestWorkspaceKey_CRUD(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	repo := "file://" + localRepo(t)
	created := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta"})
	assert.Equal(t, http.StatusCreated, created.Status, string(created.Raw))
	id := created.str("workspace", "id")
	keyring := f.opts.WorkspaceKeys
	assert.FileExists(t, filepath.Join(keyring.Dir(id), "encrypt", "id_ed25519"))

	private := func(res response) {
		t.Helper()
		assert.NotContains(t, string(res.Raw), "PRIVATE KEY")
	}

	list := f.call(t, asAdmin, http.MethodGet, "/workspaces/malta/keys", nil)
	assert.Equal(t, http.StatusOK, list.Status, string(list.Raw))
	private(list)
	assert.Equal(t, keyring.EncryptKey(id).Fingerprint, list.str("encrypt", "fingerprint"))
	assert.Len(t, list.list("keys"), 0)

	generated := f.call(t, asAdmin, http.MethodPost, "/workspaces/malta/keys/generate", api.WorkspaceKeyRequest{Type: "ed25519", Name: "deploy", Comment: "deploy"})
	assert.Equal(t, http.StatusCreated, generated.Status, string(generated.Raw))
	private(generated)
	assert.Equal(t, "id_ed25519_deploy", generated.str("key", "name"))
	assert.True(t, strings.HasPrefix(generated.str("key", "public_key"), "ssh-ed25519 "))

	granted := f.call(t, asAdmin, http.MethodPost, "/workspaces/malta/keys/generate", api.WorkspaceKeyRequest{Type: "ed25519", Name: "admin", Policies: []string{model.RoleAdmin}, Users: []string{" ops "}})
	assert.Equal(t, http.StatusCreated, granted.Status, string(granted.Raw))
	for _, name := range []string{"id_ed25519_deploy", "id_ed25519_admin"} {
		info, err := os.Stat(filepath.Join(keyring.SSH(id).Dir, name))
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	}

	// An upload: the private key of another pair.
	data, err := os.ReadFile(filepath.Join(keyring.Encrypt(id).Dir, "id_ed25519"))
	assert.NoError(t, err)
	uploaded := f.call(t, asAdmin, http.MethodPost, "/workspaces/malta/keys/upload", api.WorkspaceKeyRequest{Content: string(data), Name: "copy"})
	assert.Equal(t, http.StatusCreated, uploaded.Status, string(uploaded.Raw))
	private(uploaded)
	assert.Equal(t, "id_ed25519_copy", uploaded.str("key", "name"))
	bad := f.call(t, asAdmin, http.MethodPost, "/workspaces/malta/keys/upload", api.WorkspaceKeyRequest{Content: "not a key"})
	assert.Equal(t, http.StatusBadRequest, bad.Status)

	list = f.call(t, asAdmin, http.MethodGet, "/workspaces/malta/keys", nil)
	private(list)
	assert.Len(t, list.list("keys"), 3)
	assert.Equal(t, []any{model.RoleAdmin}, list.get("policies", "id_ed25519_admin"))
	assert.Equal(t, []any{"ops"}, list.get("users", "id_ed25519_admin"))
	assert.Nil(t, list.get("policies", "id_ed25519_deploy"))

	// A rotation keeps the grants; a PUT replaces them, none is every run.
	rotated := f.call(t, asAdmin, http.MethodPost, "/workspaces/malta/keys/generate", api.WorkspaceKeyRequest{Type: "ed25519", Name: "admin"})
	assert.Equal(t, http.StatusCreated, rotated.Status)
	list = f.call(t, asAdmin, http.MethodGet, "/workspaces/malta/keys", nil)
	assert.Equal(t, []any{model.RoleAdmin}, list.get("policies", "id_ed25519_admin"))
	put := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta/keys/id_ed25519_admin", api.WorkspaceKeyRequest{Policies: []string{model.RoleRunner}})
	assert.Equal(t, http.StatusOK, put.Status, string(put.Raw))
	list = f.call(t, asAdmin, http.MethodGet, "/workspaces/malta/keys", nil)
	assert.Equal(t, []any{model.RoleRunner}, list.get("policies", "id_ed25519_admin"))
	assert.Nil(t, list.get("users", "id_ed25519_admin"))
	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPut, "/workspaces/malta/keys/id_ed25519_admin", api.WorkspaceKeyRequest{Policies: []string{"root"}}).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodPut, "/workspaces/malta/keys/id_ed25519_nope", api.WorkspaceKeyRequest{}).Status)

	// The encryption key rotates, and no route deletes it.
	before := keyring.EncryptKey(id).Fingerprint
	encrypt := f.call(t, asAdmin, http.MethodPost, "/workspaces/malta/keys/generate", api.WorkspaceKeyRequest{Encrypt: true, Comment: "rotated"})
	assert.Equal(t, http.StatusCreated, encrypt.Status, string(encrypt.Raw))
	private(encrypt)
	assert.NotEqual(t, before, encrypt.str("key", "fingerprint"))
	assert.Equal(t, encrypt.str("key", "fingerprint"), keyring.EncryptKey(id).Fingerprint)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodDelete, "/workspaces/malta/keys/id_ed25519", nil).Status)

	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/workspaces/malta/keys/id_ed25519_admin", nil).Status)
	list = f.call(t, asAdmin, http.MethodGet, "/workspaces/malta/keys", nil)
	assert.Nil(t, list.get("policies", "id_ed25519_admin"))
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodDelete, "/workspaces/malta/keys/id_ed25519_admin", nil).Status)

	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/workspaces/malta", nil).Status)
	assert.NoDirExists(t, keyring.Dir(id))
}
