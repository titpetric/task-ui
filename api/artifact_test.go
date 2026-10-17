package api_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestArtifact_ListAndDownload(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	id := f.start(t, asOps, "infra", "artifact")
	assert.Equal(t, model.StateOK, f.finished(t, id).State)

	var list response
	assert.Eventually(t, func() bool {
		list = f.call(t, asViewer, http.MethodGet, "/artifacts?run="+id, nil)
		return len(list.list("artifacts")) == 1
	}, 5e9, 2e7)
	artifact := list.list("artifacts")[0].(map[string]any)
	assert.Equal(t, "out.txt", artifact["name"])
	assert.Equal(t, "infra", artifact["folder"])
	assert.Equal(t, float64(6), artifact["byte_size"])

	download := f.call(t, asViewer, http.MethodGet, "/artifacts/"+artifact["id"].(string), nil)
	assert.Equal(t, http.StatusOK, download.Status)
	assert.Equal(t, "hello\n", string(download.Raw))
	assert.Equal(t, "6", download.Headers.Get("Content-Length"))
	assert.Contains(t, download.Headers.Get("Content-Disposition"), `filename="out.txt"`)

	byFolder := f.call(t, asOps, http.MethodGet, "/artifacts?folder=infra&job=artifact&limit=10", nil)
	assert.Len(t, byFolder.list("artifacts"), 1)
}

func TestArtifact_Visibility(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	hidden := &model.Artifact{RunID: model.NewID(), Folder: "private", Job: "deploy", Name: "secret.txt", Path: "x/secret.txt", ByteSize: 1}
	assert.NoError(t, f.opts.Artifacts.Create(context.Background(), hidden))

	all := f.call(t, asOps, http.MethodGet, "/artifacts", nil)
	assert.Len(t, all.list("artifacts"), 1)

	limited := f.call(t, asLimited, http.MethodGet, "/artifacts", nil)
	assert.Len(t, limited.list("artifacts"), 0)

	res := f.call(t, asLimited, http.MethodGet, "/artifacts/"+hidden.ID, nil)
	assert.Equal(t, http.StatusNotFound, res.Status)

	gone := f.call(t, asOps, http.MethodGet, "/artifacts/"+hidden.ID, nil)
	assert.Equal(t, http.StatusNotFound, gone.Status)
	assert.Equal(t, "the file is gone", gone.str("error"))
}

func TestArtifact_Delete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	id := f.start(t, asOps, "infra", "artifact")
	f.finished(t, id)
	var list response
	assert.Eventually(t, func() bool {
		list = f.call(t, asOps, http.MethodGet, "/artifacts?run="+id, nil)
		return len(list.list("artifacts")) == 1
	}, 5e9, 2e7)
	artifactID := list.list("artifacts")[0].(map[string]any)["id"].(string)

	denied := f.call(t, asOps, http.MethodDelete, "/artifacts/"+artifactID, nil)
	assert.Equal(t, http.StatusForbidden, denied.Status)

	deleted := f.call(t, asAdmin, http.MethodDelete, "/artifacts/"+artifactID, nil)
	assert.Equal(t, http.StatusNoContent, deleted.Status)

	after := f.call(t, asAdmin, http.MethodGet, "/artifacts/"+artifactID, nil)
	assert.Equal(t, http.StatusNotFound, after.Status)

	again := f.call(t, asAdmin, http.MethodDelete, "/artifacts/"+artifactID, nil)
	assert.Equal(t, http.StatusNotFound, again.Status)
}
