package frontend_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
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
