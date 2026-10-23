//go:build integration

package runner_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

func TestManager_Artifacts(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	session, err := f.manager.Start(ctx, f.shellJob("art", `mkdir -p "$TASK_UI_ARTIFACTS/sub"; echo hi > "$TASK_UI_ARTIFACTS/out.txt"; echo x > "$TASK_UI_ARTIFACTS/sub/data.json"`, false))
	assert.NoError(t, err)
	_, status := collect(t, session.Attach(true))
	assert.Equal(t, model.StateOK, status.State)

	artifacts, err := f.files.List(ctx, storage.ArtifactFilter{RunID: session.ID()})
	assert.NoError(t, err)
	assert.Len(t, artifacts, 2)
	byName := map[string]model.Artifact{}
	for _, artifact := range artifacts {
		byName[artifact.Name] = artifact
	}
	out := byName["out.txt"]
	assert.Equal(t, int64(3), out.ByteSize)
	assert.Contains(t, out.ContentType, "text/plain")
	assert.Equal(t, "application/json", byName["sub/data.json"].ContentType)

	file, err := f.manager.OpenArtifact(&out)
	assert.NoError(t, err)
	defer file.Close()
	assert.FileExists(t, f.manager.ArtifactPath(&out))

	// A run that leaves nothing has no directory.
	empty, err := f.manager.Start(ctx, f.shellJob("none", "true", false))
	assert.NoError(t, err)
	collect(t, empty.Attach(true))
	assert.NoDirExists(t, filepath.Join(f.manager.ArtifactPath(&model.Artifact{Path: empty.ID()})))
}

// With an artifact store, a run's files go into its tree: a second run that
// writes the same name replaces it there, each artifact opens the version
// its run wrote, and the run's own directory goes. Past Keep commits the
// history squashes and the older versions go with it.
func TestManager_ArtifactStore(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	root := t.TempDir()
	store := &artifactstore.Store{Dir: filepath.Join(root, "tree"), Keep: 2}
	m := runner.New(runner.Options{
		Runs: f.runs, Artifacts: f.files, Recordings: &recording.Store{Root: t.TempDir()},
		ArtifactRoot: root, ArtifactStore: store, Grace: time.Second, KillTimeout: 5 * time.Second, Keep: time.Minute,
	})
	t.Cleanup(func() { m.Shutdown(context.Background()) })

	write := func(id, script string) string {
		session, err := m.Start(ctx, f.shellJob(id, script, false))
		assert.NoError(t, err)
		_, status := collect(t, session.Attach(true))
		assert.Equal(t, model.StateOK, status.State)
		return session.ID()
	}
	first := write("job1", `echo one > "$TASK_UI_ARTIFACTS/job1.txt"`)
	write("job2", `mkdir -p "$TASK_UI_ARTIFACTS/reports"; echo two > "$TASK_UI_ARTIFACTS/reports/job2.txt"`)

	artifacts, err := f.files.List(ctx, storage.ArtifactFilter{RunID: first})
	assert.NoError(t, err)
	assert.Len(t, artifacts, 1)
	one := artifacts[0]
	assert.Equal(t, "job1.txt", one.Path)
	assert.NotEqual(t, "", one.CommitID)
	assert.NoDirExists(t, filepath.Join(root, first))

	file, err := m.OpenArtifact(&one)
	assert.NoError(t, err)
	data, _ := io.ReadAll(file)
	file.Close()
	assert.Equal(t, "one\n", string(data), "the first run's version")

	again := write("job1", `echo once more > "$TASK_UI_ARTIFACTS/job1.txt"`)
	latest, _ := os.ReadFile(filepath.Join(root, "tree", "job1.txt"))
	assert.Equal(t, "once more\n", string(latest))
	assert.FileExists(t, filepath.Join(root, "tree", "reports", "job2.txt"))

	// The third commit is past Keep: the history squashes, and the first
	// version of job1.txt goes with its artifact.
	all, _ := f.files.List(ctx, storage.ArtifactFilter{})
	assert.Len(t, all, 2)
	for _, artifact := range all {
		assert.NotEqual(t, first, artifact.RunID)
	}
	current, _ := f.files.List(ctx, storage.ArtifactFilter{RunID: again})
	file, err = m.OpenArtifact(&current[0])
	assert.NoError(t, err)
	data, _ = io.ReadAll(file)
	file.Close()
	assert.Equal(t, "once more\n", string(data))

	// Empty deletes the tree and every record.
	assert.NoError(t, m.EmptyArtifacts(ctx))
	all, _ = f.files.List(ctx, storage.ArtifactFilter{})
	assert.Len(t, all, 0)
	assert.NoDirExists(t, filepath.Join(root, "tree"))
}

// A user writes a file into the tree with no run: the record names the
// user and the commit, and a second write keeps the first in the history.
func TestManager_WriteArtifact(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	store := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree")}
	m := runner.New(runner.Options{Runs: f.runs, Artifacts: f.files, ArtifactStore: store, Grace: time.Second, Keep: time.Minute})
	t.Cleanup(func() { m.Shutdown(context.Background()) })

	first, err := m.WriteArtifact(ctx, "notes/plan.md", []byte("one"), "ops")
	assert.NoError(t, err)
	assert.Equal(t, "ops", first.UserName)
	assert.Equal(t, "", first.RunID)
	assert.Equal(t, "text/markdown; charset=utf-8", first.ContentType)
	_, err = m.WriteArtifact(ctx, "notes/plan.md", []byte("two"), "ops")
	assert.NoError(t, err)
	file, err := m.OpenArtifact(first)
	assert.NoError(t, err)
	data, _ := io.ReadAll(file)
	file.Close()
	assert.Equal(t, "one", string(data))
	all, _ := f.files.List(ctx, storage.ArtifactFilter{Path: "notes/plan.md"})
	assert.Len(t, all, 2)

	_, err = m.WriteArtifact(ctx, "../out", []byte("x"), "ops")
	assert.ErrorIs(t, err, artifactstore.ErrPath)
	_, err = runner.New(runner.Options{}).WriteArtifact(ctx, "x", nil, "ops")
	assert.ErrorIs(t, err, runner.ErrNotFound)
}
