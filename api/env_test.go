package api_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestEnv_List(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.call(t, asViewer, http.MethodGet, "/env/infra", nil)
	assert.Equal(t, http.StatusOK, res.Status)
	assert.Len(t, res.list("files"), 1)
	file := res.list("files")[0].(map[string]any)
	assert.Equal(t, ".env", file["path"])
	assert.Equal(t, false, file["exists"])
	assert.False(t, contains(res.Raw, "content"))

	hidden := f.call(t, asLimited, http.MethodGet, "/env/private", nil)
	assert.Equal(t, http.StatusNotFound, hidden.Status)
}

func TestEnv_GetAndPut(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)

	denied := f.call(t, asOps, http.MethodGet, "/env/infra/.env", nil)
	assert.Equal(t, http.StatusForbidden, denied.Status)

	f.call(t, asAdmin, http.MethodPost, "/secrets", api.SecretRequest{Scope: "infra", Name: "TOKEN", Content: "t"})

	before := f.call(t, asAdmin, http.MethodGet, "/env/infra/.env", nil)
	assert.Equal(t, http.StatusOK, before.Status, string(before.Raw))
	var view api.EnvView
	assert.NoError(t, json.Unmarshal(before.Raw, &view))
	assert.Equal(t, "missing", view.State())
	assert.Equal(t, []string{"TOKEN"}, view.Secrets)

	res := f.call(t, asAdmin, http.MethodPut, "/env/infra/.env", api.EnvRequest{Content: "A=1\r\nB=two"})
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	assert.Equal(t, "managed", view.State())
	assert.Equal(t, "A=1\nB=two\n", view.Content)
	assert.Equal(t, []string{"A", "B"}, view.Keys)

	// The file in the folder carries the secret, the managed content does not.
	generated, err := os.ReadFile(filepath.Join(f.root, "infra", ".env"))
	assert.NoError(t, err)
	assert.Contains(t, string(generated), "A=1")
	assert.Contains(t, string(generated), "TOKEN=t")

	after := f.call(t, asAdmin, http.MethodGet, "/env/infra/.env", nil)
	assert.Equal(t, "A=1\nB=two\n", after.str("content"))
}

func TestEnv_Refused(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	escape := f.call(t, asAdmin, http.MethodPut, "/env/infra/../x", api.EnvRequest{Content: "A=1"})
	assert.Equal(t, http.StatusBadRequest, escape.Status, string(escape.Raw))

	folder := f.call(t, asAdmin, http.MethodPut, "/env/nowhere/.env", api.EnvRequest{Content: "A=1"})
	assert.Equal(t, http.StatusNotFound, folder.Status)

	missing := f.call(t, asAdmin, http.MethodGet, "/env/nowhere/.env", nil)
	assert.Equal(t, http.StatusNotFound, missing.Status)

	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPut, "/env/infra/.env", nil).Status)
}
