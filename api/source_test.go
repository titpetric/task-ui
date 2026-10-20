package api_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/internal/assert"
)

// The API serves a folder's source files and the files they include,
// coloured, to whoever sees the folder, and no other file.
func TestSource(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	infra := f.opts.Catalogue.Find("infra")
	assert.NoError(t, os.WriteFile(filepath.Join(infra.Dir, "atkins.yml"), []byte("jobs:\n  plan: echo <plan>\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(infra.Dir, "vars.yml"), []byte("region: eu\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(infra.Dir, ".env"), []byte("TOKEN=secret\n"), 0o644))
	infra.Sources = append(infra.Sources, &detector.Source{Kind: "atkins", File: "atkins.yml"})
	infra.Includes = []string{"vars.yml"}

	res := f.call(t, asViewer, http.MethodGet, "/sources/infra?file=atkins.yml", nil)
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	assert.Equal(t, "atkins.yml", res.str("file"))
	assert.Equal(t, "jobs:\n  plan: echo <plan>\n", res.str("content"))
	assert.Contains(t, res.str("html"), `<span class="k">jobs</span>:`)
	assert.Contains(t, res.str("html"), `&lt;plan&gt;`)

	included := f.call(t, asViewer, http.MethodGet, "/sources/infra?file=vars.yml", nil)
	assert.Equal(t, http.StatusOK, included.Status)
	assert.Equal(t, "region: eu\n", included.str("content"))

	// Only the files the folder lists: not its env file, nor a path out.
	for _, file := range []string{".env", "../private/scripts", "/etc/passwd", "missing.yml", ""} {
		assert.Equal(t, http.StatusNotFound, f.call(t, asViewer, http.MethodGet, "/sources/infra?file="+file, nil).Status, file)
	}
	// Nor a folder the user does not see.
	assert.Equal(t, http.StatusNotFound, f.call(t, asLimited, http.MethodGet, "/sources/private?file=scripts", nil).Status)
}

// The source decodes into its view: the file, verbatim and coloured.
func TestSourceView(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	infra := f.opts.Catalogue.Find("infra")
	assert.NoError(t, os.WriteFile(filepath.Join(infra.Dir, "Taskfile.yml"), []byte("tasks:\n  plan: {}\n"), 0o644))
	infra.Sources = append(infra.Sources, &detector.Source{Kind: "task", File: "Taskfile.yml"})

	res := f.call(t, asViewer, http.MethodGet, "/sources/infra?file=Taskfile.yml", nil)
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	var view api.SourceView
	assert.NoError(t, json.Unmarshal(res.Raw, &view))
	assert.Equal(t, "Taskfile.yml", view.File)
	assert.Equal(t, "tasks:\n  plan: {}\n", view.Content)
	assert.Contains(t, string(view.HTML), "tasks")
}
