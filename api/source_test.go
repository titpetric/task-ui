package api_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/internal/assert"
)

// A folder's source files and the files they include are served,
// coloured, to whoever sees the folder; no other file is.
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
