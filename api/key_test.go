package api_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
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
	assert.Equal(t, "id_ed25519", res.str("name"))
	assert.Equal(t, "deploy", res.str("comment"))
	assert.Contains(t, res.str("public_key"), "ssh-ed25519")

	list := f.call(t, asAdmin, http.MethodGet, "/keys", nil)
	assert.Len(t, list.list("keys"), 1)

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
	assert.Equal(t, key.Fingerprint, res.str("fingerprint"))
	assert.Equal(t, "uploaded", res.str("comment"))

	bad := f.call(t, asAdmin, http.MethodPost, "/keys/upload", api.KeyRequest{Content: "not a key"})
	assert.Equal(t, http.StatusBadRequest, bad.Status)
}

func TestKey_Delete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	created := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "ed25519"})
	assert.Equal(t, http.StatusCreated, created.Status)

	res := f.call(t, asAdmin, http.MethodDelete, "/keys/"+created.str("name"), nil)
	assert.Equal(t, http.StatusNoContent, res.Status)
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/keys", nil).list("keys"), 0)

	missing := f.call(t, asAdmin, http.MethodDelete, "/keys/id_rsa", nil)
	assert.Equal(t, http.StatusNotFound, missing.Status)
}
