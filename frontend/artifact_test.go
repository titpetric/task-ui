package frontend_test

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestArtifacts(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Artifacts.Create(ctx, &model.Artifact{RunID: model.NewID(), Folder: "infra", Job: "plan", Name: "plan.txt", Path: "a/plan.txt", ByteSize: 10}))
	assert.NoError(t, f.opts.Artifacts.Create(ctx, &model.Artifact{RunID: model.NewID(), Folder: "private", Job: "deploy", Name: "deploy.log", Path: "b/deploy.log", ByteSize: 20}))

	body := f.page(t, asAdmin, "/artifacts")
	assert.Contains(t, body, "<h2>Artifacts</h2>")
	assert.Contains(t, body, "plan.txt")
	assert.Contains(t, body, "deploy.log")
	assert.Contains(t, body, `data-delete="/api/artifacts/`)

	limited := f.page(t, asLimited, "/artifacts")
	assert.Contains(t, limited, "plan.txt")
	assert.NotContains(t, limited, "deploy.log")
	assert.NotContains(t, limited, "data-delete")

	filtered := f.page(t, asAdmin, "/artifacts?folder=private&job=deploy")
	assert.NotContains(t, filtered, "plan.txt")
	assert.Contains(t, filtered, `<option value="private" selected>`)

	empty := f.page(t, asAdmin, "/artifacts?folder=nowhere")
	assert.Contains(t, empty, "No artifacts yet")
}

// store puts files into the fixture's artifact tree as a run of a folder
// would, and records them.
func store(t *testing.T, f *fixture, folderID string, files map[string]string) {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		assert.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	run := model.NewID()
	commit, err := f.opts.ArtifactStore.Ingest(ctx, dir, folderID+" run", "ops")
	assert.NoError(t, err)
	for rel, content := range files {
		assert.NoError(t, f.opts.Artifacts.Create(ctx, &model.Artifact{RunID: run, Folder: folderID, Job: "build", Name: rel, Path: rel, CommitID: commit, ByteSize: int64(len(content))}))
	}
}

// The tree lists the latest files a user may see, links each, opens an
// image in the image window, and shows the space it all takes.
func TestArtifacts_Tree(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	store(t, f, "infra", map[string]string{"job1.txt": "one", "reports/chart.svg": "<svg xmlns='http://www.w3.org/2000/svg'/>"})
	store(t, f, "private", map[string]string{"job2.txt": "two"})

	body := f.page(t, asAdmin, "/artifacts")
	assert.Contains(t, body, `href="/artifacts/files/job1.txt"`)
	assert.Contains(t, body, `href="/artifacts/files/job2.txt"`)
	assert.Contains(t, body, `href="/artifacts?path=reports"`)
	assert.Contains(t, body, `class="spacebar"`)
	assert.Contains(t, body, `data-delete="/api/artifacts"`)

	limited := f.page(t, asLimited, "/artifacts")
	assert.Contains(t, limited, "job1.txt")
	assert.NotContains(t, limited, `/artifacts/files/job2.txt`)
	assert.NotContains(t, limited, `data-delete="/api/artifacts"`)

	reports := f.page(t, asLimited, "/artifacts?path=reports")
	assert.Contains(t, reports, `data-preview="/artifacts/files/reports/chart.svg"`)
	assert.Contains(t, reports, `<a href="/artifacts">Artifacts</a>`)
}

// The page serves a file inline in a sandbox, or saves it with
// ?download; one the user may not see is not there, and a directory goes
// to its listing.
func TestArtifacts_File(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	store(t, f, "infra", map[string]string{"reports/chart.svg": "<svg/>"})
	store(t, f, "private", map[string]string{"secret.txt": "x"})

	res := f.get(t, asOps, "/artifacts/files/reports/chart.svg")
	assert.Equal(t, http.StatusOK, res.Status)
	assert.Equal(t, "<svg/>", res.Body)
	assert.Equal(t, "image/svg+xml", res.Headers.Get("Content-Type"))
	assert.Contains(t, res.Headers.Get("Content-Security-Policy"), "sandbox")
	assert.Contains(t, res.Headers.Get("Content-Disposition"), "inline")

	saved := f.get(t, asOps, "/artifacts/files/reports/chart.svg?download")
	assert.Contains(t, saved.Headers.Get("Content-Disposition"), "attachment")

	assert.Equal(t, http.StatusNotFound, f.get(t, asLimited, "/artifacts/files/secret.txt").Status)
	assert.Equal(t, http.StatusOK, f.get(t, asAdmin, "/artifacts/files/secret.txt").Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/artifacts/files/missing.txt").Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/artifacts/files/../frontend_test.go").Status)

	dir := f.get(t, asAdmin, "/artifacts/files/reports")
	assert.Equal(t, http.StatusSeeOther, dir.Status)
	assert.Equal(t, "/artifacts?path=reports", dir.Location)
}

// A folder, or the whole tree, downloads as a .tgz of the latest files the
// user may see.
func TestArtifacts_Archive(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	store(t, f, "infra", map[string]string{"reports/chart.svg": "<svg/>", "job1.txt": "one"})
	store(t, f, "private", map[string]string{"secret.txt": "x"})

	names := func(user, path string) []string {
		res := f.get(t, user, path)
		assert.Equal(t, http.StatusOK, res.Status, path)
		assert.Equal(t, "application/gzip", res.Headers.Get("Content-Type"))
		zr, err := gzip.NewReader(strings.NewReader(res.Body))
		assert.NoError(t, err)
		tr := tar.NewReader(zr)
		var got []string
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			assert.NoError(t, err)
			got = append(got, header.Name)
		}
		sort.Strings(got)
		return got
	}
	assert.Equal(t, []string{"reports/chart.svg"}, names(asLimited, "/artifacts/reports.tgz"))
	assert.Equal(t, []string{"artifacts/job1.txt", "artifacts/reports/chart.svg"}, names(asLimited, "/artifacts.tgz"))
	assert.Equal(t, []string{"artifacts/job1.txt", "artifacts/reports/chart.svg", "artifacts/secret.txt"}, names(asAdmin, "/artifacts.tgz"))
	assert.Contains(t, f.get(t, asAdmin, "/artifacts/reports.tgz").Headers.Get("Content-Disposition"), `filename="reports.tgz"`)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/artifacts/missing.tgz").Status)

	page := f.page(t, asAdmin, "/artifacts")
	assert.Contains(t, page, `data-copy="/artifacts/reports.tgz"`)
	assert.Contains(t, page, `href="/artifacts/files/job1.txt?download"`)
	assert.Contains(t, page, `Delete all</button>`)
	assert.Contains(t, page, `href="/artifacts.tgz" download`)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/artifacts/reports").Status, "only a .tgz")
}
