package folder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
)

// An item is a name, whether a directory, its size and its time.
func TestItem(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "chart.SVG"), []byte("<svg/>"), 0o644))
	var items []folder.Item = folder.Listing(dir)
	assert.Len(t, items, 2)
	assert.Equal(t, "sub", items[0].Name)
	assert.True(t, items[0].Dir)
	assert.Equal(t, int64(0), items[0].Size)
	assert.False(t, items[0].ModTime.IsZero())
	assert.Equal(t, int64(6), items[1].Size)
	assert.False(t, items[1].Exec)
	assert.False(t, items[1].Sealed)
}
