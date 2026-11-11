package diskusage_test

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/internal/assert"
)

// put writes n bytes to rel under dir, making the directories.
func put(t *testing.T, dir, rel string, n int) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	assert.NoError(t, os.WriteFile(path, make([]byte, n), 0o644))
}

// The inventory sums each item's paths, files and directories alike,
// counts a path that is not there as nothing, totals the rows, and
// keeps the table for the TTL.
func TestInventory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	put(t, dir, "task-ui.db", 100)
	put(t, dir, "task-ui.db-wal", 20)
	put(t, dir, "recordings/a/cast.json", 7)
	put(t, dir, "recordings/b/cast.json", 3)
	assert.NoError(t, os.Symlink(filepath.Join(dir, "task-ui.db"), filepath.Join(dir, "recordings", "link")))

	calls := 0
	inv := &diskusage.Inventory{TTL: time.Hour, Wait: 5 * time.Second, Items: func() []diskusage.Item {
		calls++
		return []diskusage.Item{
			{Name: "task-ui database", Where: filepath.Join(dir, "task-ui.db"), Paths: []string{filepath.Join(dir, "task-ui.db"), filepath.Join(dir, "task-ui.db-wal"), filepath.Join(dir, "task-ui.db-shm")}},
			{Name: "Recordings", Where: filepath.Join(dir, "recordings"), Paths: []string{filepath.Join(dir, "recordings")}},
			{Name: "Artifacts", Where: filepath.Join(dir, "artifacts"), Paths: []string{filepath.Join(dir, "artifacts")}},
		}
	}}
	var table diskusage.Table = inv.Table()
	assert.False(t, table.Pending)
	assert.Equal(t, []diskusage.Row{
		{Name: "task-ui database", Where: filepath.Join(dir, "task-ui.db"), Size: 120, Files: 2},
		{Name: "Recordings", Where: filepath.Join(dir, "recordings"), Size: 10, Files: 2},
		{Name: "Artifacts", Where: filepath.Join(dir, "artifacts")},
	}, table.Rows, "the shm file and the artifacts are not there; the link does not count")
	assert.Equal(t, diskusage.Row{Name: "Total", Size: 130, Files: 4}, table.Total)

	put(t, dir, "recordings/c/cast.json", 1000)
	assert.Equal(t, int64(130), inv.Table().Total.Size, "kept for the TTL")
	assert.Equal(t, 1, calls)
}

// A measure that outlasts Wait answers pending: rows without numbers
// the first time, the last table marked pending after that.
func TestInventory_Pending(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	put(t, dir, "a.txt", 5)
	release := make(chan struct{})
	var calls atomic.Int32
	inv := &diskusage.Inventory{Wait: 20 * time.Millisecond, Items: func() []diskusage.Item {
		if calls.Add(1) > 1 {
			<-release
		}
		return []diskusage.Item{{Name: "Files", Where: dir, Paths: []string{dir}}}
	}}
	table := inv.Table()
	assert.False(t, table.Pending)
	assert.Equal(t, int64(5), table.Rows[0].Size)

	// TTL zero: the next call measures again, and Items blocks it.
	table = inv.Table()
	assert.True(t, table.Pending)
	assert.Equal(t, int64(5), table.Rows[0].Size, "the last measure shows meanwhile")
	assert.Equal(t, int64(5), table.Total.Size, "its total too")
	close(release)

	assert.Eventually(t, func() bool { return !inv.Table().Pending }, time.Second, 5*time.Millisecond)
}

// Before any measure, a slow one answers rows named and pending.
func TestInventory_PendingFirst(t *testing.T) {
	t.Parallel()
	release := make(chan struct{})
	var calls atomic.Int32
	inv := &diskusage.Inventory{TTL: time.Hour, Wait: 10 * time.Millisecond, Items: func() []diskusage.Item {
		if calls.Add(1) == 1 {
			<-release
		}
		return []diskusage.Item{{Name: "Files", Where: "/nowhere", Paths: []string{"/nowhere"}}}
	}}
	table := inv.Table()
	assert.True(t, table.Pending)
	assert.Equal(t, []diskusage.Row{{Name: "Files", Where: "/nowhere", Pending: true}}, table.Rows)
	assert.True(t, table.Total.Pending)
	close(release)
	assert.Eventually(t, func() bool { return !inv.Table().Pending }, time.Second, 5*time.Millisecond)
	assert.Equal(t, diskusage.Row{Name: "Files", Where: "/nowhere"}, inv.Table().Rows[0], "a path that is not there")
}

// Count adds up the regular files at a path, a file or a directory, and
// finds nothing where there is nothing.
func TestCount(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	put(t, dir, "a.txt", 3)
	put(t, dir, "sub/b.txt", 4)
	size, files := diskusage.Count(dir)
	assert.Equal(t, int64(7), size)
	assert.Equal(t, int64(2), files)
	size, files = diskusage.Count(filepath.Join(dir, "a.txt"))
	assert.Equal(t, int64(3), size)
	assert.Equal(t, int64(1), files)
	size, files = diskusage.Count(filepath.Join(dir, "nothing"))
	assert.Equal(t, int64(0), size)
	assert.Equal(t, int64(0), files)
}
