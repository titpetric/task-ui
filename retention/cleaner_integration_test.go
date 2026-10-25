//go:build integration

package retention_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/retention"
	"github.com/titpetric/task-ui/storage"
)

func newCleaner(t *testing.T) *fixture {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "task-ui.db")
	db, err := storage.OpenAndMigrate(context.Background(), "file:"+path)
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	now := time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)
	return &fixture{
		cleaner: &retention.Cleaner{
			DB: db, DBPath: path, Values: storage.NewSettingStorage(db), Runs: storage.NewRunStorage(db), Summaries: storage.NewRunSummaryStorage(db),
			Notes: storage.NewNoteStorage(db), Flows: storage.NewWorkflowRunStorage(db),
			Artifacts: storage.NewArtifactStorage(db), Store: &artifactstore.Store{Dir: filepath.Join(dir, "tree")},
			SecretReads: storage.NewSecretReadStorage(db), Tokens: storage.NewUserAuthTokenStorage(db), Sessions: storage.NewUserSessionStorage(db),
			RecordingsDir: filepath.Join(dir, "recordings"),
			Now:           func() time.Time { return now },
		},
		now: now,
		ctx: context.Background(),
	}
}

// keep returns the defaults with the kinds named kept for ever, so a
// test sees one rule at work.
func keep(kinds ...string) retention.Settings {
	s := retention.Defaults()
	for _, kind := range retention.Kinds {
		s[kind] = retention.Rule{Age: retention.Never}
	}
	for i := 0; i+1 < len(kinds); i += 2 {
		s[kinds[i]] = retention.Rule{Age: kinds[i+1]}
	}
	return s
}

// A dry run prints a line per area in the table's order, the fixed rows
// too, and no key read log.
func TestCleaner_DryRunPrintsEveryArea(t *testing.T) {
	t.Parallel()
	f := newCleaner(t)
	var out bytes.Buffer
	assert.NoError(t, f.cleaner.Clean(f.ctx, true, &out))
	lines := out.String()
	assert.Contains(t, lines, "Run recordings (keep for 31 days, no size limit): would remove 0 recordings, 0 B; 0 recordings, 0 B left\n")
	assert.Contains(t, lines, "Run history (keep for 31 days, no size limit): would remove 0 runs, 0 B; 0 runs, 0 B left\n")
	assert.Contains(t, lines, "Prompt threads (keep for 31 days): would remove 0 threads; 0 threads left\n")
	assert.Contains(t, lines, "Workflow run history (keep for 31 days): would remove 0 runs; 0 runs left\n")
	assert.Contains(t, lines, "Artifact history (keep for 31 days, no size limit): would remove 0 versions, 0 B; 0 versions, 0 B left\n")
	assert.Contains(t, lines, "Secret read log (keep for 31 days): would remove 0 reads; 0 reads left\n")
	assert.Contains(t, lines, "Expired sign-ins (removed every cleanup): would remove 0 rows; 0 rows left\n")
	assert.Contains(t, lines, "Tool history (keep for 31 days): 0 databases, 0 B\n")
	assert.NotContains(t, lines, "Key read log")
	for i, area := range retention.Areas {
		if i > 0 {
			assert.Less(t, bytes.Index(out.Bytes(), []byte(retention.Areas[i-1].Title)), bytes.Index(out.Bytes(), []byte(area.Title)), area.Title)
		}
	}
}

// Prompt threads whose last reply is older than the age go with their
// replies; a thread with a fresh reply, one with an open instruction,
// and the board's notes stay, and so do the runs.
func TestCleaner_Prompts(t *testing.T) {
	t.Parallel()
	f := newCleaner(t)
	run := f.recorded(t, 40, 100, model.StateOK)
	old := f.thread(t, 40, run.ID, "")
	revived := f.thread(t, 40, "", "")
	reply := &model.Note{Channel: model.ChannelNotes, ParentID: revived.ID, UserName: "claude", Body: "more", WorkspaceID: revived.WorkspaceID}
	assert.NoError(t, f.cleaner.Notes.Create(f.ctx, reply))
	f.backdate(t, reply.ID, f.now.Add(-2*24*time.Hour))
	open := f.thread(t, 40, "", model.NoteOpen)
	board := &model.Note{Channel: model.ChannelNotes, UserName: "ops", Body: "remember the milk"}
	assert.NoError(t, f.cleaner.Notes.Create(f.ctx, board))
	f.backdate(t, board.ID, f.now.Add(-400*24*time.Hour))
	assert.NoError(t, f.cleaner.Save(f.ctx, keep(retention.Prompts, retention.Month)))

	usage, _ := f.cleaner.Usage(f.ctx)
	assert.Equal(t, 3, usage[retention.Prompts].Count, "the board's note is no thread")

	var out bytes.Buffer
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Contains(t, out.String(), "Prompt threads (keep for 31 days): removed 1 thread; 2 threads left")
	_, err := f.cleaner.Notes.Get(f.ctx, old.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	for _, note := range []*model.Note{revived, reply, open, board} {
		_, err := f.cleaner.Notes.Get(f.ctx, note.ID)
		assert.NoError(t, err, note.Body)
	}
	_, err = f.cleaner.Runs.Get(f.ctx, run.ID)
	assert.NoError(t, err, "the run stays")
}

// Workflow runs older than the age go with their steps and the threads
// of their prompts; one in progress stays, and the workflow is not a
// row of this table.
func TestCleaner_WorkflowRuns(t *testing.T) {
	t.Parallel()
	f := newCleaner(t)
	flows := f.cleaner.Flows
	old := &model.WorkflowRun{WorkflowID: "w1", Name: "ship"}
	assert.NoError(t, flows.Create(f.ctx, old, []model.WorkflowStep{{Folder: "infra", Job: "build"}}))
	old.State = model.StateOK
	assert.NoError(t, flows.Update(f.ctx, old))
	f.backdateFlow(t, old.ID, f.now.Add(-40*24*time.Hour))
	live := &model.WorkflowRun{WorkflowID: "w1", Name: "ship"}
	assert.NoError(t, flows.Create(f.ctx, live, []model.WorkflowStep{{Folder: "infra", Job: "build"}}))
	f.backdateFlow(t, live.ID, f.now.Add(-40*24*time.Hour))
	fresh := &model.WorkflowRun{WorkflowID: "w1", Name: "ship"}
	assert.NoError(t, flows.Create(f.ctx, fresh, []model.WorkflowStep{{Folder: "infra", Job: "build"}}))
	thread := f.thread(t, 1, "", "")
	f.about(t, thread.ID, old.ID)
	assert.NoError(t, f.cleaner.Save(f.ctx, keep(retention.WorkflowRuns, retention.Month)))

	var out bytes.Buffer
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Contains(t, out.String(), "Workflow run history (keep for 31 days): removed 1 run; 2 runs left")
	_, err := flows.Get(f.ctx, old.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	steps, err := flows.Steps(f.ctx, old.ID)
	assert.NoError(t, err)
	assert.Len(t, steps, 0)
	_, err = f.cleaner.Notes.Get(f.ctx, thread.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound, "the thread of its prompt goes")
	for _, run := range []*model.WorkflowRun{live, fresh} {
		_, err := flows.Get(f.ctx, run.ID)
		assert.NoError(t, err)
		steps, _ := flows.Steps(f.ctx, run.ID)
		assert.Len(t, steps, 1)
	}
}

// Earlier versions older than the rule go with their commits; the
// latest version of every file stays and reads the same, the history
// after the cut keeps its versions, and the cleaner vacuums the database.
func TestCleaner_Artifacts(t *testing.T) {
	t.Parallel()
	f := newCleaner(t)
	f.version(t, 20, "a.txt", "a one")
	f.version(t, 19, "keep.txt", "kept all along")
	f.version(t, 18, "a.txt", "a two")
	f.version(t, 2, "a.txt", "a three")
	f.version(t, 1, "a.txt", "a four")
	assert.NoError(t, f.cleaner.Save(f.ctx, keep(retention.Artifacts, retention.Week)))

	var out bytes.Buffer
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Contains(t, out.String(), "Artifact history (keep for 7 days, no size limit): removed 2 versions, 10 B; 1 version, 7 B left")
	assert.Contains(t, out.String(), "Database: vacuumed, ")

	latest, err := f.cleaner.Artifacts.Latest(f.ctx)
	assert.NoError(t, err)
	assert.Len(t, latest, 2)
	for _, record := range latest {
		data, err := f.cleaner.Store.Show(f.ctx, record.CommitID, record.Path)
		assert.NoError(t, err, record.Path)
		want := map[string]string{"a.txt": "a four", "keep.txt": "kept all along"}[record.Path]
		assert.Equal(t, want, string(data))
	}
	history, _ := f.cleaner.Artifacts.History(f.ctx)
	assert.Len(t, history, 1)
	data, err := f.cleaner.Store.Show(f.ctx, history[0].CommitID, "a.txt")
	assert.NoError(t, err)
	assert.Equal(t, "a three", string(data))
	commits, _ := f.cleaner.Store.Commits(f.ctx)
	assert.Len(t, commits, 2)

	// Never with no limit keeps the rest.
	assert.NoError(t, f.cleaner.Save(f.ctx, keep()))
	out.Reset()
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Contains(t, out.String(), "Artifact history (keep for ever, no size limit): removed 0 versions, 0 B; 1 version, 7 B left")
}

// The secret read log loses the reads older than its age.
func TestCleaner_SecretReads(t *testing.T) {
	t.Parallel()
	f := newCleaner(t)
	var reads []model.SecretRead
	for i := 0; i < 5; i++ {
		reads = append(reads, model.SecretRead{SecretID: "s1", Name: "TOKEN", RunID: model.NewID(), Folder: "infra", Job: "build", UserName: "ops"})
	}
	assert.NoError(t, f.cleaner.SecretReads.Record(f.ctx, reads))
	for _, read := range reads[:3] {
		f.backdateRead(t, read.ID, f.now.Add(-40*24*time.Hour))
	}
	assert.NoError(t, f.cleaner.Save(f.ctx, keep(retention.SecretReads, retention.Month)))
	var out bytes.Buffer
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Contains(t, out.String(), "Secret read log (keep for 31 days): removed 3 reads; 2 reads left")
	left, _ := f.cleaner.SecretReads.All(f.ctx)
	assert.Len(t, left, 2)
	assert.Equal(t, reads[3].ID, left[0].ID)
}

// Expired tokens and sessions go every cleanup, whatever the rules; the
// live ones and the tokens without an expiry stay.
func TestCleaner_Signins(t *testing.T) {
	t.Parallel()
	f := newCleaner(t)
	assert.NoError(t, f.cleaner.Tokens.Set(f.ctx, "u1", "cli", "h1"))
	_, err := f.cleaner.Tokens.SetUntil(f.ctx, "u1", "mcp", "h2", f.now.Add(-time.Hour))
	assert.NoError(t, err)
	_, err = f.cleaner.Tokens.SetUntil(f.ctx, "u2", "mcp", "h3", f.now.Add(time.Hour))
	assert.NoError(t, err)
	_, err = f.cleaner.Sessions.Create(f.ctx, "u1", "s1", f.now.Add(-time.Minute))
	assert.NoError(t, err)
	_, err = f.cleaner.Sessions.Create(f.ctx, "u1", "s2", f.now.Add(time.Minute))
	assert.NoError(t, err)
	assert.NoError(t, f.cleaner.Save(f.ctx, keep()))

	usage, _ := f.cleaner.Usage(f.ctx)
	assert.Equal(t, 2, usage[retention.Signins].Count)
	var out bytes.Buffer
	assert.NoError(t, f.cleaner.Clean(f.ctx, true, &out))
	assert.Contains(t, out.String(), "Expired sign-ins (removed every cleanup): would remove 2 rows; 0 rows left")
	_, err = f.cleaner.Tokens.Find(f.ctx, "h2")
	assert.NoError(t, err, "a dry run removes nothing")

	out.Reset()
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Contains(t, out.String(), "Expired sign-ins (removed every cleanup): removed 2 rows; 0 rows left")
	_, err = f.cleaner.Tokens.Find(f.ctx, "h2")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = f.cleaner.Sessions.Find(f.ctx, "s1")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	for _, hash := range []string{"h1", "h3"} {
		_, err := f.cleaner.Tokens.Find(f.ctx, hash)
		assert.NoError(t, err, hash)
	}
	_, err = f.cleaner.Sessions.Find(f.ctx, "s2")
	assert.NoError(t, err)
}
