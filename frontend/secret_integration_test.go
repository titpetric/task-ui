//go:build integration

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
	assert.Contains(t, body, `href="/settings/secrets/edit?scope=infra&amp;name=DB_PASS"`)
	assert.Contains(t, body, `data-delete="/api/secrets/`)
}

func TestSecretForms(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "infra", Name: "DB_PASS", Content: "x"}))

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

// The secrets page is about secrets: no fingerprint and no key to
// rotate, since the instance's one encryption key encrypts every
// workspace's. Without it Add secret is off; with it every workspace
// group adds a secret of its scope. ?workspace= shows one group, and
// ?rekeyed= says what a rotation of the instance key wrote again.
func TestSecrets_Encryption(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "acme", Title: "Acme"}))

	noKey := f.page(t, asAdmin, "/settings/secrets")
	assert.Contains(t, noKey, `disabled title="The instance has no encryption key yet"`)
	assert.NotContains(t, noKey, `href="/settings/secrets/new`)

	f.withKey(t)
	body := f.page(t, asAdmin, "/settings/secrets")
	assert.NotContains(t, body, "SHA256:")
	assert.NotContains(t, body, `data-api="/api/secrets/rekey"`)
	assert.NotContains(t, body, "/keys/generate")
	assert.Contains(t, body, `<section class="block secret-group" id="ws-internal">`)
	assert.Contains(t, body, `<a class="btn btn-sm" href="/settings/secrets/new?scope=workspace:internal">`)
	assert.Contains(t, body, `<section class="block secret-group" id="ws-acme">`)
	assert.Contains(t, body, `<a class="btn btn-sm" href="/settings/secrets/new?scope=workspace:acme">`)
	assert.Equal(t, 0, strings.Count(body, `class="more-menu"`), "no secret, no menu")

	one := f.page(t, asAdmin, "/settings/secrets?workspace=acme")
	assert.Contains(t, one, `id="ws-acme"`)
	assert.NotContains(t, one, `id="ws-internal"`)
	assert.Contains(t, one, `href="/settings/secrets/new?scope=workspace:acme"`)

	assert.Contains(t, f.page(t, asAdmin, "/settings/secrets?rekeyed=3"), "Encryption key rotated. Re-encrypted 3 secrets.")
	assert.Contains(t, f.page(t, asAdmin, "/settings/secrets?rekeyed=1"), "Re-encrypted 1 secret.")
	assert.Contains(t, f.page(t, asAdmin, "/settings/keys?rekeyed=2&what=integration+setting"), "Re-encrypted 2 integration settings.")

	form := f.page(t, asAdmin, "/settings/secrets/new?scope=workspace:acme")
	assert.Contains(t, form, `<optgroup label="Acme">`)
	assert.Contains(t, form, `<option value="workspace:acme" selected>every folder in Acme</option>`)
	assert.Contains(t, form, `<optgroup label="Internal">`)
	assert.Contains(t, form, `<option value="workspace:internal">every folder in Internal</option>`)
	assert.Contains(t, f.page(t, asAdmin, "/settings/secrets/new"), `<option value="workspace:internal" selected>every folder in Internal</option>`, "the root first")
}

// A secret's row counts its reads and names the last run; the count
// links to its read log, paged, and a run's page names the secrets it
// got for whoever may see secrets, never a value.
func TestSecrets_Reads(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "infra", Name: "DB_PASS", Content: "hunter2secret"}))
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "infra", Name: "UNREAD", Content: "x"}))
	rows, err := f.opts.Secrets.Reaching(ctx, "infra")
	assert.NoError(t, err)
	var id string
	for _, row := range rows {
		if row.Name == "DB_PASS" {
			id = row.ID
		}
	}
	var last *model.Run
	for i := 0; i < 52; i++ {
		last = f.run(t, "infra", "plan", "ops", model.StateOK)
		assert.NoError(t, f.opts.SecretReads.Record(ctx, []model.SecretRead{{SecretID: id, Name: "DB_PASS", RunID: last.ID, Folder: "infra", Job: "plan", UserName: "ops"}}))
	}

	body := f.page(t, asAdmin, "/settings/secrets")
	assert.Contains(t, body, `<th>Reads</th><th>Last read</th>`)
	assert.Contains(t, body, `<a href="/settings/secrets/`+id+`/reads" title="The runs that got DB_PASS">52</a>`)
	assert.Contains(t, body, `<a href="/runs/`+last.ID+`">infra/plan</a>`)
	assert.Contains(t, body, `<span class="faint">never</span>`)

	log := f.page(t, asAdmin, "/settings/secrets/"+id+"/reads")
	assert.Contains(t, log, "<h2>Reads of DB_PASS")
	assert.Equal(t, 50, strings.Count(log, `<td class="key">infra/plan</td>`))
	assert.Contains(t, log, `<a class="faint mono" href="/runs/`+last.ID+`" title="`+last.ID+`">`+last.ID[:8]+`</a>`)
	assert.Contains(t, log, `rel="next">Next</a>`)
	assert.NotContains(t, log, "hunter2secret")
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/secrets/"+model.NewID()+"/reads").Status)
	assert.Equal(t, http.StatusForbidden, f.get(t, asOps, "/settings/secrets/"+id+"/reads").Status)

	// A run's page says nothing of its environment, to an admin too.
	run := f.page(t, asAdmin, "/runs/"+last.ID)
	assert.NotContains(t, run, "<th>Secrets</th>")
	assert.NotContains(t, run, "DB_PASS")
	assert.NotContains(t, run, "hunter2secret")
}

// The secrets page of a workspace lists its secrets alone and edits
// them in place; no value shows, and a workspace unknown is not found.
func TestSecrets_Workspace(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "infra", Name: "DB_PASS", Content: "hunter2secret"}))

	body := f.page(t, asAdmin, "/settings/workspaces/internal/secrets")
	assert.Contains(t, body, "<h2>Secrets of internal</h2>")
	assert.Contains(t, body, `<td class="key">DB_PASS</td>`)
	assert.Contains(t, body, `data-fill-scope="infra" data-fill-name="DB_PASS"`)
	assert.Contains(t, body, `id="secret-add-dialog"`)
	assert.NotContains(t, body, "hunter2secret")

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/workspaces/nope/secrets").Status)
}
