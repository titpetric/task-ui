package api_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// version writes a file into the artifact tree as a run of the folder,
// with its record, and returns the commit.
func (f *fixture) version(t *testing.T, folder, rel, content string) string {
	t.Helper()
	src := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Dir(filepath.Join(src, rel)), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(src, rel), []byte(content), 0o644))
	commit, err := f.opts.ArtifactStore.Ingest(context.Background(), src, folder+" run", "ops")
	assert.NoError(t, err)
	assert.NoError(t, f.opts.Artifacts.Create(context.Background(), &model.Artifact{RunID: "r-" + commit[:7], Folder: folder, Job: "build", Name: filepath.Base(rel), Path: rel, ByteSize: int64(len(content)), CommitID: commit}))
	return commit
}

// The artifact window gets a file's latest version, markdown rendered,
// and every version; ?version= picks an older one.
func TestArtifactFile_View(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	first := f.version(t, "infra", "reports/r.md", "# One\n\n![plot](plot.png)\n")
	f.version(t, "infra", "reports/r.md", "# Two\n")
	f.version(t, "private", "secret.txt", "hidden")

	latest := f.call(t, asLimited, http.MethodGet, "/artifacts/files/reports/r.md", nil)
	assert.Equal(t, http.StatusOK, latest.Status, string(latest.Raw))
	assert.Equal(t, "markdown", latest.Body["kind"])
	assert.Contains(t, latest.Body["html"].(string), `<h1 id="two">Two</h1>`)
	assert.Len(t, latest.list("versions"), 2)

	older := f.call(t, asLimited, http.MethodGet, "/artifacts/files/reports/r.md?version="+first, nil)
	assert.Equal(t, first, older.Body["version"])
	assert.Contains(t, older.Body["html"].(string), `<img src="/artifacts/files/reports/plot.png" alt="plot" />`)

	assert.Equal(t, http.StatusNotFound, f.call(t, asLimited, http.MethodGet, "/artifacts/files/secret.txt", nil).Status)
	text := f.call(t, asAdmin, http.MethodGet, "/artifacts/files/secret.txt", nil)
	assert.Equal(t, "text", text.Body["kind"])
	assert.Equal(t, "hidden", text.Body["text"])
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodGet, "/artifacts/files/nothing.txt", nil).Status)
}

// An admin deletes a file from the tree; its versions stay viewable.
func TestArtifactFile_Delete(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.version(t, "infra", "out.txt", "kept in history")
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodDelete, "/artifacts/files/out.txt", nil).Status)
	assert.Equal(t, http.StatusNoContent, f.call(t, asAdmin, http.MethodDelete, "/artifacts/files/out.txt", nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodDelete, "/artifacts/files/out.txt", nil).Status)
	path, _ := f.opts.ArtifactStore.Path("out.txt")
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err))
	gone := f.call(t, asAdmin, http.MethodGet, "/artifacts/files/out.txt", nil)
	assert.Equal(t, http.StatusOK, gone.Status)
	assert.Equal(t, "kept in history", gone.Body["text"])
}
