package frontend_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestSecrets(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	assert.NoError(t, f.opts.Secrets.Set(context.Background(), &model.Secret{Scope: "infra", Name: "DB_PASS", Content: "hunter2secret"}))

	body := f.page(t, asAdmin, "/settings/secrets")
	assert.Contains(t, body, "<h2>Secrets</h2>")
	assert.Contains(t, body, "DB_PASS")
	assert.NotContains(t, body, "characters")
	assert.Contains(t, body, `<td class="scope faint">infra</td>`)
	assert.NotContains(t, body, "hunter2secret")
	assert.NotContains(t, body, "hun")
	assert.Contains(t, body, `href="/settings/secrets/edit?scope=infra&name=DB_PASS"`)
	assert.Contains(t, body, `data-delete="/api/secrets/`)
}

func TestSecretForms(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	root, err := f.opts.Workspaces.GetByFolder(ctx, model.Default)
	assert.NoError(t, err)
	_, err = f.opts.WorkspaceKeys.Ensure(root.ID, "task-ui _")
	assert.NoError(t, err)
	assert.NoError(t, f.opts.Secrets.Set(context.Background(), &model.Secret{Scope: "infra", Name: "DB_PASS", Content: "x"}))

	add := f.page(t, asAdmin, "/settings/secrets/new?scope=infra")
	assert.Contains(t, add, "<h2>Add secret</h2>")
	assert.Contains(t, add, `<option value="infra" selected>`)
	assert.NotContains(t, add, "readonly")

	edit := f.page(t, asAdmin, "/settings/secrets/edit?scope=infra&name=DB_PASS")
	assert.Contains(t, edit, "<h2>Replace DB_PASS</h2>")
	assert.Contains(t, edit, `value="DB_PASS"`)
	assert.Contains(t, edit, "readonly")
	assert.NotContains(t, edit, ">x</textarea>")

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/secrets/edit?scope=all&name=NOPE").Status)
}

// The secrets page is about secrets: no fingerprint. A workspace with
// an encryption key has Add secret and a rotation behind its menu; one
// without says to add a key, its Add secret off. ?workspace= shows one,
// and ?rekeyed= says what a rotation wrote again.
func TestSecrets_Encryption(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	f.withKey(t)
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "acme", Title: "Acme"}))
	acme, err := f.opts.Workspaces.GetByFolder(ctx, "acme")
	assert.NoError(t, err)
	_, err = f.opts.WorkspaceKeys.Ensure(acme.ID, "task-ui acme")
	assert.NoError(t, err)

	body := f.page(t, asAdmin, "/settings/secrets")
	assert.NotContains(t, body, "SHA256:")
	assert.NotContains(t, body, `data-api="/api/secrets/rekey"`)
	assert.Contains(t, body, `<section class="block secret-group" id="ws-_">`)
	assert.Contains(t, body, `Add an encryption key to store secrets in _. <a href="/settings/keys#ws-_">Keys</a>`)
	assert.Contains(t, body, `<button class="btn btn-sm" type="button" disabled title="Add an encryption key to _ first">Add secret</button>`)
	assert.Contains(t, body, `href="/settings/secrets/new?scope=workspace%3aacme"`)
	assert.Contains(t, body, `data-api="/api/workspaces/acme/keys/generate"`)
	assert.Contains(t, body, "Rotate encryption key")
	assert.Equal(t, 1, strings.Count(body, `class="more-menu"`))

	one := f.page(t, asAdmin, "/settings/secrets?workspace=acme")
	assert.Contains(t, one, `id="ws-acme"`)
	assert.NotContains(t, one, `id="ws-_"`)
	assert.Contains(t, one, "&workspace=acme#ws-acme")

	assert.Contains(t, f.page(t, asAdmin, "/settings/secrets?rekeyed=3"), "Encryption key rotated. Re-encrypted 3 secrets.")
	assert.Contains(t, f.page(t, asAdmin, "/settings/secrets?rekeyed=1"), "Re-encrypted 1 secret.")
	assert.Contains(t, f.page(t, asAdmin, "/settings/keys?rekeyed=2&what=integration+setting"), "Re-encrypted 2 integration settings.")

	form := f.page(t, asAdmin, "/settings/secrets/new?scope=workspace:acme")
	assert.Contains(t, form, `<optgroup label="Acme">`)
	assert.Contains(t, form, `<option value="workspace:acme" selected>every project in acme</option>`)
	assert.Contains(t, form, `<optgroup label="Default: no encryption key" disabled>`)
	assert.Contains(t, form, `<option value="workspace:_" disabled>every project in _</option>`)
	assert.Contains(t, f.page(t, asAdmin, "/settings/secrets/new"), `<option value="workspace:acme" selected>`, "a workspace without a key is passed over")
}
