//go:build integration

package diskusage_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/internal/assert"
)

// In a checkout, what .gitignore leaves out counts as Ignored, outside
// the total.
func TestMeasure_Ignored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	assert.NoError(t, exec.Command("git", "init", "-q", dir).Run())
	assert.NoError(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("build/\n*.out\n"), 0o644))
	write(t, dir, "main.go", 100)
	write(t, dir, "build/bin/app", 1000)
	write(t, dir, "sub/cover.out", 200)

	u := diskusage.Measure(dir)
	assert.Equal(t, int64(113), u.Files)
	assert.Equal(t, int64(1200), u.Ignored)
	assert.Equal(t, u.Files+u.History, u.Total())
}

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
