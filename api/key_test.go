package api_test

import (
	"context"
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
	assert.Equal(t, "id_ed25519", res.str("key", "name"))
	assert.Equal(t, "deploy", res.str("key", "comment"))
	assert.Contains(t, res.str("key", "public_key"), "ssh-ed25519")
	assert.Equal(t, float64(0), res.get("rekeyed"))

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
	assert.Equal(t, key.Fingerprint, res.str("key", "fingerprint"))
	assert.Equal(t, "uploaded", res.str("key", "comment"))

	bad := f.call(t, asAdmin, http.MethodPost, "/keys/upload", api.KeyRequest{Content: "not a key"})
	assert.Equal(t, http.StatusBadRequest, bad.Status)
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

// A new key writes the secrets again for it; the last key that opens
// them stays until another is there; a backup goes after a rekey.
func TestKey_Secrets(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	assert.Equal(t, http.StatusOK, f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Name: "TOKEN", Content: "hunter2"}).Status)

	rotated := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "ed25519", Comment: "rotated"})
	assert.Equal(t, http.StatusCreated, rotated.Status, string(rotated.Raw))
	assert.Equal(t, float64(1), rotated.get("rekeyed"))
	assert.Equal(t, "id_ed25519", rotated.str("key", "name"))

	last := f.call(t, asAdmin, http.MethodDelete, "/keys/id_ed25519", nil)
	assert.Equal(t, http.StatusConflict, last.Status)
	assert.Contains(t, last.str("error"), "generate or upload another")

	var backup string
	for _, key := range f.opts.Keys.List() {
		if key.Backup {
			backup = key.Name
		}
	}
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/keys/"+backup, nil).Status)
	secrets, err := f.opts.Secrets.For(context.Background(), "infra")
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", secrets["TOKEN"])

	// Another key, then the first can go: the secrets follow.
	other := f.call(t, asAdmin, http.MethodPost, "/keys/generate", api.KeyRequest{Type: "rsa"})
	assert.Equal(t, http.StatusCreated, other.Status)
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/keys/id_ed25519", nil).Status)
	secrets, err = f.opts.Secrets.For(context.Background(), "infra")
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", secrets["TOKEN"])
}
