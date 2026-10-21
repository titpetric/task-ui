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
	assert.Contains(t, body, `title="The integrations' settings are encrypted with this key"`)
	// One badge: the ECDSA key does not encrypt.
	assert.Equal(t, 1, strings.Count(body, "integrations</span>"))
}

// A project that is a checkout with keys in its .ssh/ folder gets a
// warning: runs do not load them, upload and map them instead. A folder
// that is not a checkout gets none.
func TestKeys_ProjectFolders(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	dir := filepath.Join(f.infra.Dir, ".ssh")
	_, err := (&sshkey.Store{Dir: dir}).Generate("rsa", "deploy")
	assert.NoError(t, err)

	plain := f.page(t, asAdmin, "/settings/keys")
	assert.NotContains(t, plain, dir)

	assert.NoError(t, os.MkdirAll(filepath.Join(f.infra.Dir, ".git"), 0o755))
	body := f.page(t, asAdmin, "/settings/keys")
	assert.Contains(t, body, "<b>infra</b> has private keys in <b>"+dir+"</b>: id_rsa. They are not loaded into its runs.")
	assert.NotContains(t, body, `data-copy="ssh-rsa`)
	assert.NotContains(t, body, `data-delete="/api/keys/id_rsa"`)
}

// A key shows the roles whose runs get it, and its edit form has a box
// per role, checked for those; generate and upload show the same boxes.
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

// The forms show a name and a box per user; a user's SSH keys page lists
// the keys given to them or their role, to copy.
func TestKeys_ForAUser(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	_, err := f.opts.Keys.Generate("ed25519", "instance")
	assert.NoError(t, err)
	mine, err := f.opts.Keys.GenerateAs("ed25519", "ops", "ops key")
	assert.NoError(t, err)
	assert.NoError(t, f.opts.KeyPolicies.SetUsers(ctx, "id_ed25519_ops", []string{asOps}))
	assert.NoError(t, f.opts.KeyPolicies.Set(ctx, "id_ed25519", []string{"runner"}))

	generate := f.page(t, asAdmin, "/settings/keys/generate")
	assert.Contains(t, generate, `<select id="key-for" name="users" data-type="list" data-key-for>`)
	assert.Contains(t, generate, `<option value="`+asOps+`">`+asOps+` (runner)</option>`)
	assert.NotContains(t, generate, `name="users" value="`+asOps+`" data-type="set"`)
	edit := f.page(t, asAdmin, "/settings/keys/id_ed25519_ops")
	assert.Contains(t, edit, `name="users" value="`+asOps+`" data-type="set" checked`)
	list := f.page(t, asAdmin, "/settings/keys")
	assert.Contains(t, list, `<input type="hidden" name="name" value="ops">`)
	// The policy table: ops gets its own key and, from the runner role,
	// the instance's.
	holders := list[strings.Index(list, `class="table key-holders"`):]
	assert.Contains(t, holders, `<a class="key-chip" href="/settings/keys/id_ed25519_ops" title="id_ed25519_ops, `)
	assert.Contains(t, holders, `<a class="key-chip faint" href="/settings/keys/id_ed25519" title="id_ed25519, from the runner role">instance <small>via role</small></a>`)

	ops := f.page(t, asOps, "/keys")
	assert.Contains(t, ops, "Your SSH keys")
	assert.Contains(t, ops, strings.ReplaceAll(mine.Fingerprint, "+", "&#43;"))
	assert.Contains(t, ops, "id_ed25519_ops")
	assert.Contains(t, ops, "id_ed25519</span>") // the runner role's
	assert.Contains(t, ops, `href="/keys"`)

	viewer := f.page(t, asViewer, "/keys")
	assert.Contains(t, viewer, "No key is given to you or your role")
}

// The server user's own keys, in its ~/.ssh, show apart and read-only:
// copy, no edit or delete.
func TestKeys_HomeKeys(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	assert.NotContains(t, f.page(t, asAdmin, "/settings/keys"), "Server user's keys")
	_, err := f.opts.HomeKeys.Generate("ed25519", "root@task")
	assert.NoError(t, err)
	body := f.page(t, asAdmin, "/settings/keys")
	assert.Contains(t, body, "Server user's keys <small>"+f.opts.HomeKeys.Dir+"</small>")
	assert.Contains(t, body, `data-copy="ssh-ed25519`)
	assert.NotContains(t, body, `data-delete="/api/keys/id_ed25519"`)
}
