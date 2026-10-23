package frontend_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// A workspace's configure page lists its keys: the encryption key with
// its fingerprint, rotate and upload, and the SSH keys with whom each
// goes to, every run without a grant. The key pages generate, upload and
// grant; none of them shows a private key.
func TestWorkspaceKeys(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "malta", SourceURL: "git@example.com:org/workspace.git"}))
	ws, err := f.opts.Workspaces.GetByFolder(ctx, "malta")
	assert.NoError(t, err)
	missing := f.page(t, asAdmin, "/settings/workspaces/malta")
	assert.Contains(t, missing, `<span class="badge bad">missing</span>`)
	assert.Contains(t, missing, `data-next="/settings/workspaces/malta?notice=Encryption+key+added#keys"`)
	scope := keystore.ScopeWorkspace(ws.ID)
	encrypt := f.gen(t, keystore.Spec{Name: "workspace_malta_encryption", Purpose: keystore.PurposeEncrypt, Scope: scope})
	deploy := f.gen(t, keystore.Spec{Name: "workspace_malta_deploy", Scope: scope})
	admin := f.gen(t, keystore.Spec{Name: "workspace_malta_admin", Scope: scope})
	assert.NoError(t, f.opts.Keys.Link(ctx, admin.Name, []string{model.RoleAdmin}, nil))

	body := f.page(t, asAdmin, "/settings/workspaces/malta")
	assert.Contains(t, body, `<section class="block workspace-keys" id="keys">`)
	assert.Contains(t, body, `<form class="form" data-export="/api/workspaces/malta/export">`)
	assert.Contains(t, body, `name="passphrase" autocomplete="new-password" required`)
	assert.Contains(t, body, html(encrypt.Fingerprint))
	assert.NotContains(t, body, `<span class="badge bad">missing</span>`)
	assert.Contains(t, body, `<input type="hidden" name="encrypt" value="true" data-type="bool">`)
	assert.Contains(t, body, html(deploy.Fingerprint))
	assert.Contains(t, body, `<span class="dim">every run</span>`)
	assert.Contains(t, body, `<span class="badge">admin</span>`)
	assert.Contains(t, body, `href="/settings/workspaces/malta/keys/workspace_malta_deploy_ed25519"`)
	assert.Contains(t, body, `data-delete="/api/workspaces/malta/keys/workspace_malta_deploy_ed25519"`)
	assert.NotContains(t, body, "PRIVATE KEY")

	generate := f.page(t, asAdmin, "/settings/workspaces/malta/keys/generate")
	assert.Contains(t, generate, `data-api="/api/workspaces/malta/keys/generate"`)
	assert.Contains(t, generate, "with nothing checked, every run in the workspace gets it")
	assert.Contains(t, generate, `name="policies" value="admin"`)

	upload := f.page(t, asAdmin, "/settings/workspaces/malta/keys/upload")
	assert.Contains(t, upload, `data-api="/api/workspaces/malta/keys/upload"`)
	assert.NotContains(t, upload, `name="encrypt"`)
	assert.Contains(t, upload, `name="policies"`)
	encryptUpload := f.page(t, asAdmin, "/settings/workspaces/malta/keys/upload?encrypt=1")
	assert.Contains(t, encryptUpload, "<h2>Upload encryption key</h2>")
	assert.Contains(t, encryptUpload, `name="encrypt" value="true"`)
	assert.NotContains(t, encryptUpload, `name="policies"`)

	edit := f.page(t, asAdmin, "/settings/workspaces/malta/keys/"+admin.Name)
	assert.Contains(t, edit, `data-api="/api/workspaces/malta/keys/workspace_malta_admin_ed25519" data-method="PUT"`)
	assert.Contains(t, edit, `name="policies" value="admin" data-type="set" checked>`)
	assert.Contains(t, edit, html(admin.Fingerprint))

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/workspaces/malta/keys/workspace_malta_nope_ed25519").Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/workspaces/nope/keys/generate").Status)
	assert.Equal(t, http.StatusForbidden, f.get(t, asOps, "/settings/workspaces/malta/keys/generate").Status)
}

// html is a fingerprint as a page writes it, its plus signs escaped.
func html(fingerprint string) string {
	return strings.ReplaceAll(fingerprint, "+", "&#43;")
}
