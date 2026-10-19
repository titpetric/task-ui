package frontend_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshkey"
)

func TestKeys(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	empty := f.page(t, asAdmin, "/settings/keys")
	assert.Contains(t, empty, "<h2>SSH keys</h2>")
	assert.Contains(t, empty, "No key pair yet")

	key, err := f.opts.Keys.Generate("ed25519", "deploy")
	assert.NoError(t, err)
	_, err = f.opts.Keys.Generate("ed25519", "rotated")
	assert.NoError(t, err)

	body := f.page(t, asAdmin, "/settings/keys")
	assert.Contains(t, body, "rotated")
	assert.Contains(t, body, "Backups")
	assert.Contains(t, body, strings.ReplaceAll(key.Fingerprint, "+", "&#43;"))
	assert.Contains(t, body, `data-copy="ssh-ed25519`)
	assert.Contains(t, body, `data-delete="/api/keys/id_ed25519"`)
	assert.Contains(t, body, `data-api="/api/keys/generate"`)
}

func TestKeyForms(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	generate := f.page(t, asAdmin, "/settings/keys/generate")
	assert.Contains(t, generate, "<h2>Generate key</h2>")
	assert.Contains(t, generate, f.opts.Keys.Dir)
	assert.Contains(t, generate, `<option value="ed25519">`)

	upload := f.page(t, asAdmin, "/settings/keys/upload")
	assert.Contains(t, upload, "<h2>Upload key</h2>")
	assert.Contains(t, upload, `data-read-into="up-content"`)
	assert.Contains(t, upload, `data-api="/api/keys/upload"`)
}

func TestKeys_Encrypting(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	_, err := f.opts.Keys.Generate("ed25519", "deploy")
	assert.NoError(t, err)
	_, err = f.opts.Keys.Generate("ecdsa", "curve")
	assert.NoError(t, err)

	body := f.page(t, asAdmin, "/settings/keys")
	assert.Contains(t, body, `<span class="badge">id_ed25519</span>`)
	assert.Contains(t, body, `title="The secrets are encrypted with this key"`)
	// One badge: the ECDSA key does not encrypt.
	assert.Equal(t, 1, strings.Count(body, "secrets</span>"))
}

// A project that is a checkout with keys in its .ssh/ folder lists them
// under the instance keys, to copy and nothing else: the project manages
// them. A folder that is not a checkout lists none.
func TestKeys_ProjectFolders(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	dir := filepath.Join(f.infra.Dir, ".ssh")
	key, err := (&sshkey.Store{Dir: dir}).Generate("rsa", "deploy")
	assert.NoError(t, err)

	plain := f.page(t, asAdmin, "/settings/keys")
	assert.NotContains(t, plain, dir)

	assert.NoError(t, os.MkdirAll(filepath.Join(f.infra.Dir, ".git"), 0o755))
	body := f.page(t, asAdmin, "/settings/keys")
	assert.Contains(t, body, dir+" · for the runs of infra")
	assert.Contains(t, body, strings.ReplaceAll(key.Fingerprint, "+", "&#43;"))
	assert.Contains(t, body, `data-copy="ssh-rsa`)
	assert.NotContains(t, body, `data-delete="/api/keys/id_rsa"`)
}

// A key shows the roles whose runs get it, and its edit form has a box
// per role, checked for those; generate and upload offer the same boxes.
func TestKeys_Policies(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	_, err := f.opts.Keys.Generate("ed25519", "deploy")
	assert.NoError(t, err)
	assert.NoError(t, f.opts.KeyPolicies.Set(context.Background(), "id_ed25519", []string{"runner"}))

	list := f.page(t, asAdmin, "/settings/keys")
	assert.Contains(t, list, `<span class="badge">runner</span>`)
	assert.Contains(t, list, `href="/settings/keys/id_ed25519"`)

	edit := f.page(t, asAdmin, "/settings/keys/id_ed25519")
	assert.Contains(t, edit, `data-api="/api/keys/id_ed25519" data-method="PUT"`)
	assert.Contains(t, edit, `value="runner" data-type="set" checked`)
	assert.Contains(t, edit, `value="admin" data-type="set">`)
	assert.Contains(t, edit, `value="viewer" data-type="set">`)

	missing := f.get(t, asAdmin, "/settings/keys/id_rsa")
	assert.Equal(t, http.StatusNotFound, missing.Status)

	for _, path := range []string{"/settings/keys/generate", "/settings/keys/upload"} {
		assert.Contains(t, f.page(t, asAdmin, path), `name="policies" value="runner" data-type="set"`)
	}
}
