package folder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestCatalogue(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "site"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "site", "Makefile"), []byte("build:\n\techo ok\n"), 0o644))

	var c folder.Catalogue = &folder.Discovery{Root: root, Discover: true}
	list := c.List()
	assert.Len(t, list, 1)
	assert.Equal(t, "site", list[0].ID)
	assert.Equal(t, "site", c.Find("site").ID)
	assert.Nil(t, c.Find("nowhere"))
}
