package frontend_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
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
	_, err = f.opts.WorkspaceKeys.Ensure(ws.ID, "task-ui malta")
	assert.NoError(t, err)
	encrypt := f.opts.WorkspaceKeys.EncryptKey(ws.ID)
	deploy, err := f.opts.WorkspaceKeys.SSH(ws.ID).GenerateAs("ed25519", "deploy", "deploy")
	assert.NoError(t, err)
	admin, err := f.opts.WorkspaceKeys.SSH(ws.ID).GenerateAs("ed25519", "admin", "admin")
	assert.NoError(t, err)
	assert.NoError(t, f.opts.KeyGrants.Set(ctx, ws.ID, admin.Name, []string{model.RoleAdmin}, nil))

	body := f.page(t, asAdmin, "/settings/workspaces/malta")
	assert.Contains(t, body, `<section class="block workspace-keys">`)
	assert.Contains(t, body, html(encrypt.Fingerprint))
	assert.Contains(t, body, `href="/settings/workspaces/malta/keys/upload?encrypt=1"`)
	assert.Contains(t, body, `<input type="hidden" name="encrypt" value="true" data-type="bool">`)
	assert.Contains(t, body, html(deploy.Fingerprint))
	assert.Contains(t, body, `<span class="dim">every run</span>`)
	assert.Contains(t, body, `<span class="badge">admin</span>`)
	assert.Contains(t, body, `href="/settings/workspaces/malta/keys/id_ed25519_deploy"`)
	assert.Contains(t, body, `data-delete="/api/workspaces/malta/keys/id_ed25519_deploy"`)
	assert.NotContains(t, body, "PRIVATE KEY")

	generate := f.page(t, asAdmin, "/settings/workspaces/malta/keys/generate")
	assert.Contains(t, generate, `data-api="/api/workspaces/malta/keys/generate"`)
	assert.Contains(t, generate, "none checked is every run in the workspace")
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
	assert.Contains(t, edit, `data-api="/api/workspaces/malta/keys/id_ed25519_admin" data-method="PUT"`)
	assert.Contains(t, edit, `name="policies" value="admin" data-type="set" checked>`)
	assert.Contains(t, edit, html(admin.Fingerprint))

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/workspaces/malta/keys/id_ed25519_nope").Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/workspaces/nope/keys/generate").Status)
	assert.Equal(t, http.StatusForbidden, f.get(t, asOps, "/settings/workspaces/malta/keys/generate").Status)
}

// html is a fingerprint as a page writes it, its plus signs escaped.
func html(fingerprint string) string {
	return strings.ReplaceAll(fingerprint, "+", "&#43;")
}
