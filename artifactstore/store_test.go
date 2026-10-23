package artifactstore_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/internal/assert"
)

// run makes a run's artifact directory with the files given.
func run(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		assert.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	return dir
}

func read(t *testing.T, s *artifactstore.Store, rel string) string {
	t.Helper()
	path, err := s.Path(rel)
	assert.NoError(t, err)
	data, err := os.ReadFile(path)
	assert.NoError(t, err)
	return string(data)
}

// Two jobs leave their files side by side; a later run replaces a file
// and the earlier version stays in the history.
func TestStore_Ingest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree")}

	first, err := s.Ingest(ctx, run(t, map[string]string{"job1.txt": "one", "reports/a.html": "<p>a</p>"}), "infra/job1 run 1", "alice")
	assert.NoError(t, err)
	assert.NotEqual(t, "", first)
	second, err := s.Ingest(ctx, run(t, map[string]string{"job2.txt": "two"}), "infra/job2 run 2", "bob")
	assert.NoError(t, err)
	third, err := s.Ingest(ctx, run(t, map[string]string{"job1.txt": "one, again"}), "infra/job1 run 3", "alice")
	assert.NoError(t, err)
	assert.NotEqual(t, second, third)

	assert.Equal(t, "one, again", read(t, s, "job1.txt"))
	assert.Equal(t, "two", read(t, s, "job2.txt"))
	old, err := s.Show(ctx, first, "job1.txt")
	assert.NoError(t, err)
	assert.Equal(t, "one", string(old))

	// Nothing new is the same commit.
	same, err := s.Ingest(ctx, run(t, map[string]string{"job2.txt": "two"}), "infra/job2 run 4", "bob")
	assert.NoError(t, err)
	assert.Equal(t, third, same)

	names := []string{}
	for _, item := range s.List("") {
		names = append(names, item.Name)
	}
	assert.Equal(t, []string{"reports", "job1.txt", "job2.txt"}, names)
	assert.Len(t, s.List("reports"), 1)
}

// A run cannot write into the repository, and no path leaves the tree.
func TestStore_Guards(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree")}
	_, err := s.Ingest(ctx, run(t, map[string]string{".git/config": "x", "ok.txt": "ok"}), "run", "")
	assert.NoError(t, err)
	assert.Equal(t, "ok", read(t, s, "ok.txt"))

	for _, rel := range []string{"../escape", ".git/config", "a/.git/x"} {
		_, err := s.Path(rel)
		assert.ErrorIs(t, err, artifactstore.ErrNotFound, rel)
	}
	_, err = s.Show(ctx, "not-a-commit", "ok.txt")
	assert.ErrorIs(t, err, artifactstore.ErrNotFound)
}

// Past Keep commits the history flattens to one, and the old versions go.
func TestStore_Squash(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree"), Keep: 2}

	head, err := s.Squash(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "", head, "nothing to squash yet")

	first, _ := s.Ingest(ctx, run(t, map[string]string{"a.txt": "1"}), "run 1", "")
	_, _ = s.Ingest(ctx, run(t, map[string]string{"a.txt": "2"}), "run 2", "")
	head, err = s.Squash(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "", head, "two commits are within Keep")

	_, _ = s.Ingest(ctx, run(t, map[string]string{"a.txt": "3"}), "run 3", "")
	head, err = s.Squash(ctx)
	assert.NoError(t, err)
	assert.NotEqual(t, "", head)

	latest, err := s.Show(ctx, head, "a.txt")
	assert.NoError(t, err)
	assert.Equal(t, "3", string(latest))
	_, err = s.Show(ctx, first, "a.txt")
	assert.ErrorIs(t, err, artifactstore.ErrNotFound, "the first version is gone")
}

// Empty deletes everything; the next run starts a new tree.
func TestStore_Empty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree")}
	_, _ = s.Ingest(ctx, run(t, map[string]string{"a.txt": "1"}), "run 1", "")
	assert.NoError(t, s.Empty())
	assert.Len(t, s.List(""), 0)
	_, err := s.Ingest(ctx, run(t, map[string]string{"b.txt": "2"}), "run 2", "")
	assert.NoError(t, err)
	assert.Len(t, s.List(""), 1)
}

// Remove takes a file or a directory out of the tree; the history keeps
// what it held.
func TestStore_Remove(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree")}
	_, err := s.Remove(ctx, "job1.txt", "alice")
	assert.ErrorIs(t, err, artifactstore.ErrNotFound)

	first, err := s.Ingest(ctx, run(t, map[string]string{"job1.txt": "one", "reports/a.html": "a", "reports/b.html": "b"}), "run 1", "alice")
	assert.NoError(t, err)
	gone, err := s.Remove(ctx, "job1.txt", "admin")
	assert.NoError(t, err)
	assert.NotEqual(t, first, gone)
	path, _ := s.Path("job1.txt")
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
	old, err := s.Show(ctx, first, "job1.txt")
	assert.NoError(t, err)
	assert.Equal(t, "one", string(old))

	_, err = s.Remove(ctx, "reports", "admin")
	assert.NoError(t, err)
	assert.Len(t, s.List(""), 0)
	_, err = s.Remove(ctx, "job1.txt", "admin")
	assert.ErrorIs(t, err, artifactstore.ErrNotFound)
	_, err = s.Remove(ctx, "../x", "admin")
	assert.ErrorIs(t, err, artifactstore.ErrNotFound)
	_, err = s.Remove(ctx, "", "admin")
	assert.ErrorIs(t, err, artifactstore.ErrNotFound)
}

// The Then forms call then with the commit while they hold the tree; an
// error from then is theirs.
func TestStore_IngestThen(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree"), Keep: 1}
	var got string
	commit, err := s.IngestThen(ctx, run(t, map[string]string{"a.txt": "a"}), "run 1", "ops", func(c string) error { got = c; return nil })
	assert.NoError(t, err)
	assert.Equal(t, commit, got)
	_, err = s.IngestThen(ctx, run(t, map[string]string{"a.txt": "a"}), "run 2", "ops", func(string) error { return context.Canceled })
	assert.ErrorIs(t, err, context.Canceled)
}

// SquashThen calls then with the new head once it squashed, and not
// while the history is within Keep.
func TestStore_SquashThen(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree"), Keep: 1}
	_, _ = s.Write(ctx, "a.txt", []byte("a"), "ops")
	calls := 0
	head, err := s.SquashThen(ctx, func(string) error { calls++; return nil })
	assert.NoError(t, err)
	assert.Equal(t, "", head)
	_, _ = s.Write(ctx, "a.txt", []byte("b"), "ops")
	var got string
	head, err = s.SquashThen(ctx, func(h string) error { calls++; got = h; return nil })
	assert.NoError(t, err)
	assert.Equal(t, head, got)
	assert.Equal(t, 1, calls)
}
