//go:build integration

package frontend_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// gen makes a key of the fixture's key folder.
func (f *fixture) gen(t *testing.T, spec keystore.Spec) *keystore.Key {
	t.Helper()
	key, err := f.opts.Keys.Generate(context.Background(), spec)
	assert.NoError(t, err)
	return key
}

// The Keys page lists every key with its purpose, algorithm,
// fingerprint, scope and links, the git key marked.
func TestKeys(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	empty := f.page(t, asAdmin, "/settings/keys")
	assert.Contains(t, empty, "<h2>Keys</h2>")
	assert.Contains(t, empty, "No key yet")

	key := f.gen(t, keystore.Spec{Name: "deploy", Comment: "deploy"})
	assert.NoError(t, f.opts.Keys.SetGit(context.Background(), key.Name))
	f.gen(t, keystore.Spec{Name: "task_encryption", Purpose: keystore.PurposeEncrypt})

	body := f.page(t, asAdmin, "/settings/keys")
	assert.Contains(t, body, `<tr id="key-deploy_ed25519">`)
	assert.Contains(t, body, `<span class="badge ok" title="git pulls and clones with it">git</span>`)
	assert.Contains(t, body, strings.ReplaceAll(key.Fingerprint, "+", "&#43;"))
	assert.Contains(t, body, `data-copy="ssh-ed25519`)
	assert.Contains(t, body, `data-delete="/api/keys/deploy_ed25519"`)
	assert.NotContains(t, body, `data-delete="/api/keys/task_encryption_ed25519"`, "an encryption key stays")
	assert.Contains(t, body, `<span class="badge signal">encryption</span>`)
	assert.Contains(t, body, `<span class="faint">never handed out</span>`)
}

// The forms take a name, the algorithm with its size, the purpose and
// the links; ?for=<user> starts a key of that user.
func TestKeyForms(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	generate := f.page(t, asAdmin, "/settings/keys/generate")
	assert.Contains(t, generate, "<h2>Generate key</h2>")
	assert.Contains(t, generate, f.opts.Keys.Dir)
	assert.Contains(t, generate, `<option value="ed25519" selected>ed25519</option>`)
	assert.Contains(t, generate, `<option value="rsa" data-bits="4096">RSA 4096</option>`)
	assert.Contains(t, generate, `name="purpose" value="encrypt"`)
	assert.Contains(t, generate, `name="policies" value="runner" data-type="set"`)
	assert.Contains(t, generate, `id="key-name" name="name" value="" required`)

	upload := f.page(t, asAdmin, "/settings/keys/upload")
	assert.Contains(t, upload, "<h2>Upload key</h2>")
	assert.Contains(t, upload, `data-read-into="up-content"`)
	assert.Contains(t, upload, `data-api="/api/keys/upload"`)

	forOps := f.page(t, asAdmin, "/settings/keys/generate?for="+asOps)
	assert.Contains(t, forOps, `name="name" value="user_`+asOps+`" required`)
	assert.Contains(t, forOps, `name="users" value="`+asOps+`" data-type="set" checked`)
}

// The instance's encryption key has a box of its own on Keys, and on
// the Instance page with the git key.
func TestKeys_Encryption(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	instance := f.opts.Keys.EncryptKey(context.Background(), storage.ScopeInstance)
	body := f.page(t, asAdmin, "/settings/keys")
	assert.Contains(t, body, `<tr class="key-instance" id="instance">`)
	assert.Contains(t, body, strings.ReplaceAll(instance.Fingerprint, "+", "&#43;"))
	assert.Contains(t, body, `data-api="/api/keys/encrypt/rotate"`)

	assert.NoError(t, f.opts.Keys.SetGit(context.Background(), "task_ssh_ed25519"))
	page := f.page(t, asAdmin, "/settings/instance")
	assert.Contains(t, page, `<a class="mono" href="/settings/keys/task_encryption_ed25519">task_encryption_ed25519</a>`)
	assert.Contains(t, page, `<a class="mono" href="/settings/keys/task_ssh_ed25519">task_ssh_ed25519</a>`)
	assert.Contains(t, page, "Rotate Key</button>")
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
	assert.Contains(t, body, "<b>infra</b> has private keys in <b>"+dir+"</b>: id_rsa. Its runs don't get them.")
	assert.NotContains(t, body, `data-copy="ssh-rsa`)
	assert.NotContains(t, body, `data-delete="/api/keys/id_rsa"`)
}

// A key's edit form has a box per role and user, checked for its links,
// and the git box; a missing key is not found.
func TestKeys_Links(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	key := f.gen(t, keystore.Spec{Name: "deploy"})
	assert.NoError(t, f.opts.Keys.Link(context.Background(), key.Name, []string{"runner"}, []string{asOps}))

	list := f.page(t, asAdmin, "/settings/keys")
	assert.Contains(t, list, `<span class="badge">runner</span>`)
	assert.Contains(t, list, `href="/settings/keys/deploy_ed25519"`)

	edit := f.page(t, asAdmin, "/settings/keys/deploy_ed25519")
	assert.Contains(t, edit, `data-api="/api/keys/deploy_ed25519" data-method="PUT"`)
	assert.Contains(t, edit, `value="runner" data-type="set" checked`)
	assert.Contains(t, edit, `value="admin" data-type="set">`)
	assert.Contains(t, edit, `name="users" value="`+asOps+`" data-type="set" checked`)
	assert.Contains(t, edit, `<input type="checkbox" name="git"> git pulls and clones with this key`)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/keys/missing").Status)

	// The policy table: ops gets the key by name and the runner role.
	holders := list[strings.Index(list, `class="table key-holders"`):]
	assert.Contains(t, holders, `<a class="key-chip" href="/settings/keys/deploy_ed25519" title="deploy_ed25519, `)

	ops := f.page(t, asOps, "/keys")
	assert.Contains(t, ops, "Your SSH keys")
	assert.Contains(t, ops, strings.ReplaceAll(key.Fingerprint, "+", "&#43;"))
	assert.Contains(t, ops, ">you<")
	viewer := f.page(t, asViewer, "/keys")
	assert.Contains(t, viewer, "No keys for you or your role yet")

	profile := f.page(t, asAdmin, "/users/"+asOps)
	assert.Contains(t, profile, `<span class="mono">deploy_ed25519</span>`)
	assert.Contains(t, profile, `href="/settings/keys/generate?for=`+asOps+`"`)
}

// The keys page has a row per workspace: its projects, a count of its
// secrets linked to them, its encryption key, set or missing, and its
// SSH keys, or what having none means.
func TestKeys_Workspaces(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "acme", Title: "Acme"}))
	acme, err := f.opts.Workspaces.GetByFolder(ctx, "acme")
	assert.NoError(t, err)
	scope := keystore.ScopeWorkspace(acme.ID)
	enc := f.gen(t, keystore.Spec{Name: "workspace_acme_encryption", Purpose: keystore.PurposeEncrypt, Scope: scope})
	f.gen(t, keystore.Spec{Name: "workspace_acme_deploy", Scope: scope})
	f.opts.Secrets.Workspaces = f.workspaceVaults()
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "workspace:acme", Name: "TOKEN", Content: "x"}))
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "workspace:acme", Name: "OTHER", Content: "y"}))

	body := f.page(t, asAdmin, "/settings/keys")
	assert.Contains(t, body, `<tr id="ws-acme">`)
	assert.Contains(t, body, `<a href="/settings/secrets?workspace=acme">2 secrets</a>`)
	assert.Contains(t, body, `<a class="key-chip" href="/settings/workspaces/acme/keys/workspace_acme_deploy_ed25519"`)
	assert.Contains(t, body, html(enc.Fingerprint))
	assert.Contains(t, body, `<tr id="ws-_" class="key-missing">`)
	assert.Contains(t, body, "You can't add secrets until it has one.")
	assert.Contains(t, body, "This workspace has no SSH key of its own")
	assert.Contains(t, body, `href="/settings/workspaces/_/keys/upload?encrypt=1"`)
	assert.Contains(t, body, `<div class="faint key-projects">infra, private</div>`, "the root's projects are the default workspace's")
	assert.Contains(t, body, "every run of the workspace")
}
