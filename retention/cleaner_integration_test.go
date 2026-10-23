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
			DB: db, DBPath: path, Values: storage.NewSettingStorage(db), Runs: storage.NewRunStorage(db),
			Artifacts: storage.NewArtifactStorage(db), Store: &artifactstore.Store{Dir: filepath.Join(dir, "tree")},
			SecretReads: storage.NewSecretReadStorage(db), KeyReads: storage.NewKeyReadStorage(db), RecordingsDir: filepath.Join(dir, "recordings"),
			Now: func() time.Time { return now },
		},
		now: now,
		ctx: context.Background(),
	}
}

// Recordings older than the rule go, the oldest first past the limit;
// a running run's, and a running workflow's step's, stay; the run rows
// stay; a dry run removes nothing.
func TestCleaner_Recordings(t *testing.T) {
	t.Parallel()
	f := newCleaner(t)
	old := f.recorded(t, 40, 100, model.StateOK)
	running := f.recorded(t, 40, 100, model.StateRunning)
	step := f.recorded(t, 39, 100, model.StateFailed)
	mid := f.recorded(t, 10, 100, model.StateOK)
	fresh := f.recorded(t, 1, 100, model.StateOK)
	flows := storage.NewWorkflowRunStorage(f.cleaner.DB)
	flow := &model.WorkflowRun{WorkflowID: "w1", Name: "ship"}
	assert.NoError(t, flows.Create(f.ctx, flow, []model.WorkflowStep{{Folder: "internal/tools", Job: "build"}}))
	steps, _ := flows.Steps(f.ctx, flow.ID)
	steps[0].RunID = step.ID
	assert.NoError(t, flows.UpdateStep(f.ctx, &steps[0]))

	usage, err := f.cleaner.Usage(f.ctx)
	assert.NoError(t, err)
	assert.Equal(t, 5, usage[retention.Recordings].Count)
	assert.Equal(t, int64(500), usage[retention.Recordings].Bytes)

	var out bytes.Buffer
	assert.NoError(t, f.cleaner.Clean(f.ctx, true, &out))
	assert.Contains(t, out.String(), "Run recordings (keep for 31 days, no size limit): would remove 1 item, 100 B; 4 items, 400 B left")
	assert.True(t, f.exists(old))

	out.Reset()
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Contains(t, out.String(), "Run recordings (keep for 31 days, no size limit): removed 1 item, 100 B; 4 items, 400 B left")
	assert.False(t, f.exists(old))
	assert.True(t, f.exists(running))
	assert.True(t, f.exists(step))
	row, err := f.cleaner.Runs.Get(f.ctx, old.ID)
	assert.NoError(t, err, "the run stays")
	assert.Equal(t, old.CastPath, row.CastPath)

	// A size limit takes the oldest idle ones until the rest, the busy
	// ones among them, fit.
	assert.NoError(t, f.cleaner.Save(f.ctx, retention.Settings{Recordings: retention.Rule{Age: retention.Never, Limit: 350}, Artifacts: retention.Rule{Age: retention.Never}, SecretReads: retention.Rule{Age: retention.Never}}))
	out.Reset()
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.False(t, f.exists(mid))
	assert.True(t, f.exists(fresh))
	assert.True(t, f.exists(running))
	assert.True(t, f.exists(step))
	assert.NotContains(t, out.String(), "Database: vacuumed", "no rows went")
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
	assert.NoError(t, f.cleaner.Save(f.ctx, retention.Settings{Recordings: retention.Rule{Age: retention.Month}, Artifacts: retention.Rule{Age: retention.Week}, SecretReads: retention.Rule{Age: retention.Never}}))

	var out bytes.Buffer
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Contains(t, out.String(), "Artifact history (keep for 7 days, no size limit): removed 2 items, 10 B; 1 item, 7 B left")
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
	assert.NoError(t, f.cleaner.Save(f.ctx, retention.Defaults()))
	out.Reset()
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Contains(t, out.String(), "Artifact history (keep for ever, no size limit): removed 0 items, 0 B; 1 item, 7 B left")
}

// The secret read log loses its oldest reads past the limit.
func TestCleaner_SecretReads(t *testing.T) {
	t.Parallel()
	f := newCleaner(t)
	var reads []model.SecretRead
	for i := 0; i < 5; i++ {
		reads = append(reads, model.SecretRead{SecretID: "s1", Name: "TOKEN", RunID: model.NewID(), Folder: "infra", Job: "build", UserName: "ops"})
	}
	assert.NoError(t, f.cleaner.SecretReads.Record(f.ctx, reads))
	usage, _ := f.cleaner.Usage(f.ctx)
	each := usage[retention.SecretReads].Bytes / 5
	assert.NoError(t, f.cleaner.Save(f.ctx, retention.Settings{Recordings: retention.Rule{Age: retention.Month}, Artifacts: retention.Rule{Age: retention.Never}, SecretReads: retention.Rule{Age: retention.Never, Limit: 2 * each}}))
	var out bytes.Buffer
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Contains(t, out.String(), "Secret read log (keep for ever, at most ")
	assert.Contains(t, out.String(), "removed 3 items")
	left, _ := f.cleaner.SecretReads.All(f.ctx)
	assert.Len(t, left, 2)
	assert.Equal(t, reads[3].ID, left[0].ID)
}

// The key read log keeps to its rule as the secret read log does.
func TestCleaner_KeyReads(t *testing.T) {
	t.Parallel()
	f := newCleaner(t)
	var reads []*model.KeyRead
	for i := 0; i < 4; i++ {
		read := &model.KeyRead{KeyName: "deploy_rsa4096", UserName: "claude", TokenID: "t1", Via: "mcp"}
		assert.NoError(t, f.cleaner.KeyReads.Record(f.ctx, read))
		reads = append(reads, read)
	}
	usage, _ := f.cleaner.Usage(f.ctx)
	each := usage[retention.KeyReads].Bytes / 4
	assert.NoError(t, f.cleaner.Save(f.ctx, retention.Settings{Recordings: retention.Rule{Age: retention.Month}, Artifacts: retention.Rule{Age: retention.Never}, SecretReads: retention.Rule{Age: retention.Never}, KeyReads: retention.Rule{Age: retention.Never, Limit: each}}))
	var out bytes.Buffer
	assert.NoError(t, f.cleaner.Clean(f.ctx, false, &out))
	assert.Contains(t, out.String(), "Key read log (keep for ever, at most ")
	left, _ := f.cleaner.KeyReads.All(f.ctx)
	assert.Len(t, left, 1)
	assert.Equal(t, reads[3].ID, left[0].ID)
}
