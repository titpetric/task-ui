package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/storage"
)

func TestSecret_AdminOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/secrets"},
		{http.MethodPost, "/secrets"},
		{http.MethodDelete, "/secrets/" + model.NewID()},
	} {
		res := f.call(t, asOps, c.method, c.path, map[string]any{})
		assert.Equal(t, http.StatusForbidden, res.Status, c.method+" "+c.path)
	}
}

func TestSecret_Set(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)

	res := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Name: "API_TOKEN", Content: "hunter2"})
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))

	var view api.SecretView
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	assert.Equal(t, storage.ScopeAll, view.Scope)
	assert.Equal(t, "API_TOKEN", view.Name)
	assert.True(t, model.ValidID(view.ID))
	assert.False(t, string(res.Raw) != "" && contains(res.Raw, "hunter2"))

	secrets, err := f.opts.Secrets.For(context.Background(), "infra")
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", secrets["API_TOKEN"])

	// Setting again replaces the content and keeps the row.
	again := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: " all ", Name: "API_TOKEN", Content: "changed"})
	assert.Equal(t, http.StatusOK, again.Status)
	assert.Equal(t, view.ID, again.str("id"))
	secrets, _ = f.opts.Secrets.For(context.Background(), "infra")
	assert.Equal(t, "changed", secrets["API_TOKEN"])

	scoped := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: "infra", Name: "DB_PASS", Content: "x"})
	assert.Equal(t, http.StatusOK, scoped.Status)
	assert.Equal(t, "infra", scoped.str("scope"))
}

func TestSecret_SetRefused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)

	folder := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: "nowhere", Name: "X", Content: "x"})
	assert.Equal(t, http.StatusBadRequest, folder.Status)
	assert.Contains(t, folder.str("error"), "nowhere")

	name := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Name: "bad name", Content: "x"})
	assert.Equal(t, http.StatusBadRequest, name.Status)
}

func TestSecret_ListAndDelete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)

	set := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Name: "TOKEN", Content: "hunter2"})
	f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: "infra", Name: "TOKEN", Content: "other"})

	list := f.call(t, asAdmin, http.MethodGet, "/secrets", nil)
	assert.Equal(t, http.StatusOK, list.Status)
	assert.Len(t, list.list("secrets"), 2)
	assert.False(t, contains(list.Raw, "hunter2"))
	assert.False(t, contains(list.Raw, "content"))

	deleted := f.call(t, asAdmin, http.MethodDelete, "/secrets/"+set.str("id"), nil)
	assert.Equal(t, http.StatusNoContent, deleted.Status)
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/secrets", nil).list("secrets"), 1)

	// Deleting again succeeds and does nothing.
	again := f.call(t, asAdmin, http.MethodDelete, "/secrets/"+set.str("id"), nil)
	assert.Equal(t, http.StatusNoContent, again.Status)
}

// Without a key to encrypt with the store keeps no secret; a rekey writes every
// secret again for the current keys.
func TestSecret_Encryption(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	none := f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Name: "TOKEN", Content: "x"})
	assert.Equal(t, http.StatusConflict, none.Status)
	assert.Contains(t, none.str("error"), "generate an ed25519 or RSA key")

	f.withKey(t)
	assert.Equal(t, http.StatusOK, f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Name: "TOKEN", Content: "x"}).Status)
	raw, err := f.opts.Secrets.Secrets.List(context.Background())
	assert.NoError(t, err)
	assert.True(t, secret.Encrypted(raw[0].Content))

	denied := f.call(t, asOps, http.MethodPost, "/secrets/rekey", map[string]any{})
	assert.Equal(t, http.StatusForbidden, denied.Status)
	rekeyed := f.call(t, asAdmin, http.MethodPost, "/secrets/rekey", map[string]any{})
	assert.Equal(t, http.StatusOK, rekeyed.Status, string(rekeyed.Raw))
	assert.Equal(t, float64(1), rekeyed.get("rekeyed"))
	assert.Len(t, rekeyed.list("keys"), 1)
}
