//go:build integration

package api_test

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
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
		{http.MethodPut, "/workspaces/malta/keys/workspace_malta_ssh_ed25519"},
		{http.MethodDelete, "/workspaces/malta/keys/workspace_malta_ssh_ed25519"},
	} {
		for _, user := range []string{asOps, asViewer, asLimited} {
			res := f.call(t, user, c.method, c.path, map[string]any{})
			assert.Equal(t, http.StatusForbidden, res.Status, user+" "+c.method+" "+c.path)
		}
	}
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/workspaces/nope/keys", nil).Status)
}

// A workspace has no encryption key, the instance's encrypts every
// secret, so an encryption key for one answers 400; an admin generates,
// links and removes its SSH keys, named
// workspace_<folder>_<name>_<algorithm>, and no answer carries a private
// key. Deleting the workspace takes its keys.
func TestWorkspaceKey_CRUD(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	repo := "file://" + localRepo(t)
	created := f.call(t, asAdmin, http.MethodPost, "/workspaces", api.WorkspaceRequest{URL: repo, Folder: "malta"})
	assert.Equal(t, http.StatusCreated, created.Status, string(created.Raw))
	scope := keystore.ScopeWorkspace(created.str("workspace", "id"))
	assert.Nil(t, f.opts.Keys.EncryptKey(ctx, scope))
	refused := f.call(t, asAdmin, http.MethodPost, "/workspaces/malta/keys/generate", api.WorkspaceKeyRequest{Encrypt: true, Comment: "task-ui malta"})
	assert.Equal(t, http.StatusBadRequest, refused.Status, string(refused.Raw))
	assert.Equal(t, api.ErrWorkspaceEncrypt.Error(), refused.str("error"))
	assert.Nil(t, f.opts.Keys.EncryptKey(ctx, scope))
	assert.NoFileExists(t, filepath.Join(f.opts.Keys.Dir, "workspace_malta_encryption_ed25519.key"))

	ssh := f.call(t, asAdmin, http.MethodPost, "/workspaces/malta/keys/generate", api.WorkspaceKeyRequest{Name: "build", Policies: []string{model.RoleAdmin}})
	assert.Equal(t, http.StatusCreated, ssh.Status, string(ssh.Raw))
	name := ssh.str("key", "name")
	assert.Equal(t, "workspace_malta_build_ed25519", name)
	plain := f.call(t, asAdmin, http.MethodPost, "/workspaces/malta/keys/generate", api.WorkspaceKeyRequest{Algo: "ecdsa", Bits: 256})
	assert.Equal(t, "workspace_malta_ssh_ecdsa256", plain.str("key", "name"))

	list := f.call(t, asAdmin, http.MethodGet, "/workspaces/malta/keys", nil)
	assert.Equal(t, http.StatusOK, list.Status)
	assert.Len(t, list.list("keys"), 2, "the SSH keys")
	assert.Equal(t, f.opts.Keys.Dir, list.str("dir"))
	assert.NotContains(t, string(list.Raw), "PRIVATE KEY")

	put := f.call(t, asAdmin, http.MethodPut, "/workspaces/malta/keys/"+name, api.WorkspaceKeyRequest{Users: []string{"ops"}})
	assert.Equal(t, http.StatusOK, put.Status, string(put.Raw))
	key, err := f.opts.Keys.Get(ctx, name)
	assert.NoError(t, err)
	assert.Equal(t, []string{"ops"}, key.Users)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodPut, "/workspaces/malta/keys/workspace_malta_encryption_ed25519", api.WorkspaceKeyRequest{}).Status, "no encryption key to link")

	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/workspaces/malta/keys/"+name, nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodDelete, "/workspaces/malta/keys/"+name, nil).Status)

	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/workspaces/malta", nil).Status)
	left, _ := f.opts.Keys.Of(ctx, scope, "")
	assert.Len(t, left, 0, "the workspace's keys go with it")
}
