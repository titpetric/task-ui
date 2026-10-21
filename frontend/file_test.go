package frontend_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// /files/ serves a file of a checkout to whoever sees the folder, inline
// as its type says, plain text for text without an extension, and as an
// attachment with ?download.
func TestCheckoutFile(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	assert.NoError(t, os.WriteFile(filepath.Join(f.infra.Dir, "Makefile"), []byte("all:\n\techo\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(f.infra.Dir, "page.html"), []byte("<script>alert(1)</script>"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(f.infra.Dir, "blob"), []byte("a\x00b"), 0o644))

	res := f.get(t, asViewer, "/files/infra/Makefile")
	assert.Equal(t, http.StatusOK, res.Status)
	assert.Equal(t, "text/plain; charset=utf-8", res.Headers.Get("Content-Type"))
	assert.Equal(t, `inline; filename="Makefile"`, res.Headers.Get("Content-Disposition"))
	assert.Equal(t, "all:\n\techo\n", res.Body)
	page := f.get(t, asViewer, "/files/infra/page.html")
	assert.Contains(t, page.Headers.Get("Content-Security-Policy"), "sandbox")
	assert.Equal(t, "application/octet-stream", f.get(t, asViewer, "/files/infra/blob").Headers.Get("Content-Type"))
	download := f.get(t, asViewer, "/files/infra/Makefile?download")
	assert.Equal(t, `attachment; filename="Makefile"`, download.Headers.Get("Content-Disposition"))

	assert.Equal(t, http.StatusNotFound, f.get(t, asViewer, "/files/infra").Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asViewer, "/files/infra/missing").Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asLimited, "/files/private/x").Status)
}
