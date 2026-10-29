//go:build integration

package frontend_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// sideRow cuts a folder's row out of a page's sidebar, from its
// data-folder to the end of its summary.
func sideRow(body, id string) string {
	side, _, _ := strings.Cut(body, "<main")
	_, rest, ok := strings.Cut(side, `data-folder="`+id+`"`)
	if !ok {
		return ""
	}
	row, _, _ := strings.Cut(rest, "</summary>")
	return row
}

// The sidebar lights the icon of a folder whose checkout has files
// changed or untracked under it, and says how many on a mouse-over; a
// folder outside git stays unlit, on every page, the open one too.
func TestSidebar_Dirty(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	out, err := exec.Command("git", "-c", "init.defaultBranch=main", "init", "-q", f.infra.Dir).CombinedOutput()
	assert.NoError(t, err, string(out))
	assert.NoError(t, os.WriteFile(filepath.Join(f.infra.Dir, "new.txt"), []byte("x\n"), 0o644))

	for _, page := range []string{"/f/internal/infra", "/f/internal/private"} {
		body := f.page(t, asOps, page)
		assert.Contains(t, sideRow(body, "infra"), `<svg class="ic dirty"><title>1 file changed or untracked</title>`, page)
		private := sideRow(body, "private")
		assert.Contains(t, private, `<svg class="ic">`, page)
		assert.NotContains(t, private, "dirty", page)
	}
}
