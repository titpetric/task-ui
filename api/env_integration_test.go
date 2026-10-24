//go:build integration

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

	f.withKey(t)

	// A folder lists the env files users filled, none until one is.
	empty := f.call(t, asViewer, http.MethodGet, "/env/infra", nil)
	assert.Equal(t, http.StatusOK, empty.Status)
	assert.Len(t, empty.list("files"), 0)

	f.call(t, asAdmin, http.MethodPut, "/env/infra/.env", api.EnvRequest{Content: "A=1"})
	res := f.call(t, asViewer, http.MethodGet, "/env/infra", nil)
	assert.Len(t, res.list("files"), 1)
	file := res.list("files")[0].(map[string]any)
	assert.Equal(t, ".env", file["path"])
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
	assert.Equal(t, "", view.Content)
	assert.Equal(t, []string{"TOKEN"}, view.Secrets)

	res := f.call(t, asAdmin, http.MethodPut, "/env/infra/.env", api.EnvRequest{Content: "A=1\r\nB=two"})
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	assert.Equal(t, "A=1\nB=two\n", view.Content)

	// The file in the folder carries the secret, the managed content does not.
	generated, err := os.ReadFile(filepath.Join(f.root, "infra", ".env"))
	assert.NoError(t, err)
	assert.Contains(t, string(generated), "A=1")
	assert.Contains(t, string(generated), "TOKEN=t")

	after := f.call(t, asAdmin, http.MethodGet, "/env/infra/.env", nil)
	assert.Equal(t, "A=1\nB=two\n", after.str("content"))

	// Saving it empty deletes the row and the written file.
	cleared := f.call(t, asAdmin, http.MethodPut, "/env/infra/.env", api.EnvRequest{Content: ""})
	assert.Equal(t, http.StatusOK, cleared.Status, string(cleared.Raw))
	assert.NoFileExists(t, filepath.Join(f.root, "infra", ".env"))
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/env/infra", nil).list("files"), 0)
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

// A nested folder's env files resolve under its path.
func TestEnv_Nested(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)

	put := f.call(t, asAdmin, http.MethodPut, "/env/infra/edge/.env", api.EnvRequest{Content: "EDGE=1"})
	assert.Equal(t, http.StatusOK, put.Status, string(put.Raw))
	list := f.call(t, asAdmin, http.MethodGet, "/env/infra/edge", nil)
	assert.Equal(t, http.StatusOK, list.Status, string(list.Raw))
	assert.Len(t, list.list("files"), 1)
	assert.Equal(t, "infra/edge", put.str("folder"))
	assert.FileExists(t, filepath.Join(f.root, "infra", "edge", ".env"))

	got := f.call(t, asAdmin, http.MethodGet, "/env/infra/edge/.env", nil)
	assert.Equal(t, "EDGE=1\n", got.str("content"))

	project := f.call(t, asAdmin, http.MethodPut, "/projects/infra/edge", api.ProjectRequest{Title: "Edge"})
	assert.Equal(t, http.StatusOK, project.Status, string(project.Raw))
	assert.Equal(t, "infra/edge", project.str("folder"))
	assert.Equal(t, http.StatusOK, f.call(t, asAdmin, http.MethodGet, "/projects/infra/edge", nil).Status)
}
