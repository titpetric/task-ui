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

// thread adds a prompt thread, a note with a workspace, days ago, about
// a run when the test names one, in a state.
func (f *fixture) thread(t *testing.T, days int, aboutRun, state string) *model.Note {
	t.Helper()
	note := &model.Note{Channel: model.ChannelNotes, UserName: "ops", Body: "why did it fail", WorkspaceID: model.Default, Dir: "infra", AboutRunID: aboutRun, State: state}
	assert.NoError(t, f.cleaner.Notes.Create(f.ctx, note))
	reply := &model.Note{Channel: model.ChannelNotes, ParentID: note.ID, UserName: "claude", Body: "because", WorkspaceID: model.Default, AboutRunID: aboutRun}
	assert.NoError(t, f.cleaner.Notes.Create(f.ctx, reply))
	at := f.now.Add(-time.Duration(days) * 24 * time.Hour)
	f.backdate(t, note.ID, at)
	f.backdate(t, reply.ID, at)
	return note
}

// backdate moves a note's time to at.
func (f *fixture) backdate(t *testing.T, id string, at time.Time) {
	t.Helper()
	_, err := f.cleaner.DB.ExecContext(f.ctx, "UPDATE note SET created_at = ? WHERE id = ?", at.UTC(), id)
	assert.NoError(t, err)
}

// about makes a thread one of a workflow run's prompts.
func (f *fixture) about(t *testing.T, id, workflowRunID string) {
	t.Helper()
	_, err := f.cleaner.DB.ExecContext(f.ctx, "UPDATE note SET about_workflow_run_id = ? WHERE id = ? OR parent_id = ?", workflowRunID, id, id)
	assert.NoError(t, err)
}

// backdateFlow moves a workflow run's start to at.
func (f *fixture) backdateFlow(t *testing.T, id string, at time.Time) {
	t.Helper()
	_, err := f.cleaner.DB.ExecContext(f.ctx, "UPDATE workflow_run SET started_at = ? WHERE id = ?", at.UTC(), id)
	assert.NoError(t, err)
}

// backdateRead moves a secret read's time to at.
func (f *fixture) backdateRead(t *testing.T, id string, at time.Time) {
	t.Helper()
	_, err := f.cleaner.DB.ExecContext(f.ctx, "UPDATE secret_read SET created_at = ? WHERE id = ?", at.UTC(), id)
	assert.NoError(t, err)
}

// file writes size bytes at path.
func (f *fixture) file(t *testing.T, path string, size int) {
	t.Helper()
	assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	assert.NoError(t, os.WriteFile(path, bytes.Repeat([]byte("x"), size), 0o644))
}
