//go:build integration

package frontend_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestSecrets(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "infra", Name: "DB_PASS", Content: "hunter2secret"}))
	assert.NoError(t, f.opts.Secrets.SetEnv(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "PUBLIC_URL", Content: "https://example.com"}))

	body := f.page(t, asAdmin, "/settings/secrets")
	assert.Contains(t, body, "<h2>Secrets and env</h2>")
	assert.Contains(t, body, "DB_PASS")
	assert.Contains(t, body, "PUBLIC_URL")
	assert.NotContains(t, body, "hunter2secret")
	assert.NotContains(t, body, "hun")
	assert.Contains(t, body, `<td class="scope faint">infra</td>`)
	assert.Contains(t, body, `<code>https://example.com</code>`, "an env value shows")
	assert.Contains(t, body, `href="/settings/secrets/edit?id=`)
	assert.Contains(t, body, `data-delete="/api/secrets/`)
}

func TestSecretForms(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "infra", Name: "DB_PASS", Content: "x"}))

	add := f.page(t, asAdmin, "/settings/secrets/new")
	assert.Contains(t, add, "<h2>Add variable</h2>")
	assert.Contains(t, add, `<option value="all" selected>All workspaces</option>`)
	assert.Contains(t, add, `name="is_secret" checked`)
	assert.NotContains(t, add, "readonly")

	edit := f.page(t, asAdmin, "/settings/secrets/edit?scope=infra&name=DB_PASS")
	assert.Contains(t, edit, "<h2>Edit DB_PASS</h2>")
	assert.Contains(t, edit, `value="DB_PASS"`)
	assert.NotContains(t, edit, "readonly")
	assert.NotContains(t, edit, ">x</textarea>", "a secret's value never shows")

	rows, err := f.opts.Secrets.Secrets.List(ctx)
	assert.NoError(t, err)
	id := rows[0].ID
	byID := f.page(t, asAdmin, "/settings/secrets/edit?id="+id)
	assert.Contains(t, byID, "<h2>Edit DB_PASS</h2>")

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/secrets/edit?id="+model.NewID()).Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/secrets/edit?scope=all&name=NOPE").Status)
}

// The secrets page is about variables: no fingerprint and no key to
// rotate, since the instance's one encryption key encrypts them all.
// Without it Add variable is off; with it the form offers every
// workspace and each one. ?workspace= shows what reaches one, and
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
	assert.Contains(t, body, `href="/settings/secrets/new"`)
	assert.Equal(t, 0, strings.Count(body, `class="more-menu"`), "no variable, no menu")

	assert.Contains(t, f.page(t, asAdmin, "/settings/secrets?rekeyed=3"), "Encryption key rotated. Re-encrypted 3 secrets.")
	assert.Contains(t, f.page(t, asAdmin, "/settings/secrets?rekeyed=1"), "Re-encrypted 1 secret.")
	assert.Contains(t, f.page(t, asAdmin, "/settings/keys?rekeyed=2&what=integration+setting"), "Re-encrypted 2 integration settings.")

	form := f.page(t, asAdmin, "/settings/secrets/new?scope=workspace:acme")
	assert.Contains(t, form, `<option value="all">All workspaces</option>`)
	assert.Contains(t, form, `<option value="workspace:acme" selected>Acme</option>`)
	assert.Contains(t, form, `<option value="workspace:internal">Internal</option>`)
	assert.Contains(t, f.page(t, asAdmin, "/settings/secrets/new"), `<option value="all" selected>All workspaces</option>`)
}

// A secret's row counts its reads and names the last run; the count
// links to its read log, paged, and a run's page names no variable it
// got, neither a name nor a value.
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
	assert.Contains(t, body, `<th>Reads</th>`)
	assert.Contains(t, body, `<a href="/settings/secrets/`+id+`/reads" title="The runs that got this variable">52</a>`)
	assert.Contains(t, body, `<a class="faint" href="/runs/`+last.ID+`" title="Last read `)
	assert.Contains(t, body, `: infra/plan">`)
	assert.Contains(t, body, `<span class="faint" title="No run has got this variable">never</span>`)

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

// ?workspace= shows what reaches one workspace, its own and the
// variables of every one; the old per-workspace page is gone.
func TestSecrets_Filter(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "acme", Title: "Acme"}))
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "infra", Name: "DB_PASS", Content: "hunter2secret"}))
	assert.NoError(t, f.opts.Secrets.SetEnv(ctx, &model.Secret{Scope: "workspace:acme", Name: "ACME_URL", Content: "https://acme.example"}))
	assert.NoError(t, f.opts.Secrets.SetEnv(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "EVERYWHERE", Content: "everywhere"}))

	internal := f.page(t, asAdmin, "/settings/secrets?workspace=internal")
	assert.Contains(t, internal, "DB_PASS")
	assert.Contains(t, internal, "EVERYWHERE")
	assert.NotContains(t, internal, "ACME_URL")

	acme := f.page(t, asAdmin, "/settings/secrets?workspace=acme")
	assert.Contains(t, acme, "ACME_URL")
	assert.Contains(t, acme, "EVERYWHERE")
	assert.NotContains(t, acme, "DB_PASS")

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/workspaces/internal/secrets").Status)
}

// The integrations' computed environment is read only at the end of the
// page: a credential's value is redacted, a plain one shows, and no
// integration value is editable here.
func TestSecrets_IntegrationEnv(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "integration:openrouter", Name: "token", Content: "sk-or-secret"}))

	body := f.page(t, asAdmin, "/settings/secrets")
	assert.Contains(t, body, `<section class="block" id="integration-env">`)
	assert.Contains(t, body, `<td class="key">OPENROUTER_API_KEY</td>`)
	assert.Contains(t, body, `<td class="key">ANTHROPIC_AUTH_TOKEN</td>`)
	assert.Contains(t, body, `<td class="key">ANTHROPIC_BASE_URL</td>`)
	assert.Contains(t, body, `<span class="faint mono">REDACTED</span>`)
	assert.Contains(t, body, `<code>https://openrouter.ai/api</code>`)
	assert.NotContains(t, body, "sk-or-secret")
	assert.NotContains(t, body, `data-api="/api/integrations`)
	assert.Equal(t, 3, strings.Count(body, "REDACTED"), "the key, the auth token and the token count")
}

// The count of the variables on the settings index and a workspace's
// page link the filtered secrets page.
func TestSecrets_Counts(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Workspaces.EnsureDefault(ctx))
	assert.NoError(t, f.opts.Workspaces.Create(ctx, &model.Workspace{Folder: "acme", Title: "Acme"}))
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "workspace:acme", Name: "TOKEN", Content: "x"}))

	secrets := f.page(t, asAdmin, "/settings/secrets?workspace=acme")
	assert.Contains(t, secrets, "TOKEN")
	assert.NotContains(t, secrets, `href="/settings/workspaces/malta/secrets"`)

	settings := f.page(t, asAdmin, "/settings")
	assert.Contains(t, settings, `href="/settings/secrets"`)

	workspaces := f.page(t, asAdmin, "/settings/workspaces")
	assert.Contains(t, workspaces, `href="/settings/secrets?workspace=acme"`)
}
