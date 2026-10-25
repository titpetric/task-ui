//go:build integration

package retention_test

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/retention"
)

// The tool history row measures each atkins database that is there with
// its journal, skips one that is not, and the cleanup has atkins prune
// each under the rule's age, dry run passed through, a line per
// database; without atkins it measures and says so, and an age of never
// prunes nothing.
func TestCleaner_Tooling(t *testing.T) {
	t.Parallel()
	f := newCleaner(t)
	dir := t.TempDir()
	f.file(t, filepath.Join(dir, "atkins.db"), 300)
	f.file(t, filepath.Join(dir, "atkins.db-wal"), 50)
	f.file(t, filepath.Join(dir, "other.db"), 20)
	f.cleaner.Tooling = func() []string {
		return []string{filepath.Join(dir, "atkins.db"), filepath.Join(dir, "none", "atkins.db"), filepath.Join(dir, "other.db")}
	}
	usage, err := f.cleaner.Usage(f.ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, usage[retention.Tooling].Count)
	assert.Equal(t, int64(370), usage[retention.Tooling].Bytes)

	var out bytes.Buffer
	assert.NoError(t, f.cleaner.Clean(f.ctx, true, &out))
	assert.Contains(t, out.String(), "Tool history (keep for 31 days): 2 databases, 370 B; atkins not on PATH\n")

	var calls [][]string
	f.cleaner.Atkins = func(_ context.Context, args ...string) ([]byte, error) {
		calls = append(calls, args)
		return []byte(args[2] + ": would remove 1 run, 1 command older than 31d; 2 runs, 1 command left\n"), nil
	}
	out.Reset()
	assert.NoError(t, f.cleaner.Clean(f.ctx, true, &out))
	assert.Len(t, calls, 2)
	assert.Equal(t, []string{"--prune", "--db", filepath.Join(dir, "atkins.db"), "--older-than", "31d", "--dry-run"}, calls[0])
	assert.Equal(t, filepath.Join(dir, "other.db"), calls[1][2])
	assert.Contains(t, out.String(), "Tool history (keep for 31 days): "+filepath.Join(dir, "atkins.db")+": would remove 1 run, 1 command older than 31d; 2 runs, 1 command left\n")
	assert.Contains(t, out.String(), "Tool history (keep for 31 days): "+filepath.Join(dir, "other.db")+": would remove")

	calls = nil
	assert.NoError(t, f.cleaner.Save(f.ctx, keep(retention.Tooling, retention.Week)))
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Equal(t, []string{"--prune", "--db", filepath.Join(dir, "atkins.db"), "--older-than", "7d"}, calls[0])

	calls = nil
	out.Reset()
	assert.NoError(t, f.cleaner.Save(f.ctx, keep()))
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Empty(t, calls)
	assert.Contains(t, out.String(), "Tool history (keep for ever): 2 databases, 370 B\n")

	f.cleaner.Atkins = func(context.Context, ...string) ([]byte, error) {
		return []byte("no such table"), errors.New("exit status 1")
	}
	assert.NoError(t, f.cleaner.Save(f.ctx, keep(retention.Tooling, retention.Month)))
	err = f.cleaner.Clean(f.ctx, false, &out)
	assert.ErrorContains(t, err, "Tool history: "+filepath.Join(dir, "atkins.db")+": exit status 1: no such table")
	assert.FileExists(t, filepath.Join(dir, "atkins.db"))
}
