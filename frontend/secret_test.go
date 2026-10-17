package frontend_test

import (
	"context"
	"net/http"
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
	assert.Contains(t, body, "13 characters")
	assert.NotContains(t, body, "hunter2secret")
	assert.NotContains(t, body, "hun")
	assert.Contains(t, body, `href="/settings/secrets/edit?scope=infra&name=DB_PASS"`)
	assert.Contains(t, body, `data-delete="/api/secrets/`)
}

func TestSecretForms(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
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

func TestSecrets_Encryption(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	none := f.page(t, asAdmin, "/settings/secrets")
	assert.Contains(t, none, "No key to encrypt with")
	assert.NotContains(t, none, `data-api="/api/secrets/rekey"`)

	f.withKey(t)
	body := f.page(t, asAdmin, "/settings/secrets")
	assert.Contains(t, body, "Encrypted at rest with the instance SSH key <b>id_ed25519</b>")
	assert.Contains(t, body, `data-api="/api/secrets/rekey"`)
	assert.Contains(t, f.page(t, asAdmin, "/settings/secrets/new"), "<b>id_ed25519</b>")
	assert.Contains(t, f.page(t, asAdmin, "/settings/keys"), "encrypted with: <b>id_ed25519</b>")
}
