package retention_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/retention"
)

// fixture is a cleaner over a database file, a recordings folder and an
// artifact tree, its clock at noon.
type fixture struct {
	cleaner *retention.Cleaner
	now     time.Time
	ctx     context.Context
}

// recorded adds a run that started days ago with a recording of size
// bytes, in a state.
func (f *fixture) recorded(t *testing.T, days int, size int, state string) *model.Run {
	t.Helper()
	started := f.now.Add(-time.Duration(days) * 24 * time.Hour)
	run := &model.Run{Folder: "internal/tools", Job: "build", State: state, StartedAt: &started}
	run.ID = model.NewIDAt(started)
	run.CastPath = filepath.ToSlash(filepath.Join("internal", "build", run.ID+".cast"))
	file := filepath.Join(f.cleaner.RecordingsDir, filepath.FromSlash(run.CastPath))
	assert.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
	assert.NoError(t, os.WriteFile(file, bytes.Repeat([]byte("x"), size), 0o644))
	assert.NoError(t, f.cleaner.Runs.Create(f.ctx, run))
	return run
}

func (f *fixture) exists(run *model.Run) bool {
	_, err := os.Stat(filepath.Join(f.cleaner.RecordingsDir, filepath.FromSlash(run.CastPath)))
	return err == nil
}

// version writes a file into the tree as a run would, with its record,
// days ago.
func (f *fixture) version(t *testing.T, days int, rel, content string) {
	t.Helper()
	at := f.now.Add(-time.Duration(days) * 24 * time.Hour)
	commit, err := f.cleaner.Store.Write(f.ctx, rel, []byte(content), "ops")
	assert.NoError(t, err)
	assert.NoError(t, f.cleaner.Artifacts.Create(f.ctx, &model.Artifact{RunID: model.NewID(), Folder: "infra", Job: "build", Name: rel, Path: rel, ByteSize: int64(len(content)), CommitID: commit, CreatedAt: &at}))
}
