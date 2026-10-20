package diskusage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/internal/assert"
)

// write makes a file of n bytes under dir.
func write(t *testing.T, dir, rel string, n int) {
	t.Helper()
	path := filepath.Join(dir, rel)
	assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	assert.NoError(t, os.WriteFile(path, make([]byte, n), 0o644))
}

// Measure counts the files and the history apart.
func TestMeasure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, "a.txt", 100)
	write(t, dir, "sub/b.txt", 50)
	write(t, dir, ".git/objects/x", 300)
	write(t, dir, ".github/c.yml", 10)

	u := diskusage.Measure(dir)
	assert.Equal(t, int64(160), u.Files)
	assert.Equal(t, int64(300), u.History)
	assert.Equal(t, int64(460), u.Total())

	assert.Equal(t, diskusage.Usage{}, diskusage.Measure(filepath.Join(dir, "missing")))
}

func TestUsage_Percent(t *testing.T) {
	t.Parallel()
	u := diskusage.Usage{Files: 25, History: 75}
	assert.Equal(t, 25.0, u.FilesPercent())
	assert.Equal(t, 75.0, u.HistoryPercent())
	assert.Equal(t, 0.0, diskusage.Usage{}.FilesPercent())
	assert.Equal(t, 0.0, diskusage.Usage{}.HistoryPercent())
}
