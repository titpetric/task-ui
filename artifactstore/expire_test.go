package artifactstore_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/internal/assert"
)

// Expire drops the history before a commit: the versions only the
// dropped commits held go, the tree and the later versions stay under
// new names, and a file the kept tree holds unchanged reads the same.
func TestStore_Expire(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree")}
	first, _ := s.Write(ctx, "a.txt", []byte("one"), "ops")
	_, _ = s.Write(ctx, "b.txt", []byte("bee"), "ops")
	third, _ := s.Write(ctx, "a.txt", []byte("two"), "ops")
	fourth, _ := s.Write(ctx, "a.txt", []byte("three"), "ops")
	commits, err := s.Commits(ctx)
	assert.NoError(t, err)
	assert.Len(t, commits, 4)

	var got artifactstore.Rewrite
	assert.NoError(t, s.Expire(ctx, third, func(r artifactstore.Rewrite) error {
		got = r
		assert.True(t, r.Same(commits[1], "b.txt"), "b.txt lives on in the root")
		assert.False(t, r.Same(first, "a.txt"), "the first a.txt went")
		return nil
	}))
	assert.Equal(t, []string{first, commits[1]}, got.Dropped)
	assert.Len(t, got.Renamed, 2)
	assert.Equal(t, got.Renamed[third], got.Root)

	commits, _ = s.Commits(ctx)
	assert.Equal(t, []string{got.Root, got.Renamed[fourth]}, commits)
	old, err := s.Show(ctx, got.Root, "a.txt")
	assert.NoError(t, err)
	assert.Equal(t, "two", string(old))
	assert.Equal(t, "three", read(t, s, "a.txt"))
	_, err = s.Show(ctx, first, "a.txt")
	assert.ErrorIs(t, err, artifactstore.ErrNotFound)

	assert.NoError(t, s.Expire(ctx, got.Root, nil), "the first commit changes nothing")
	assert.ErrorIs(t, s.Expire(ctx, "0123456789abcdef", nil), artifactstore.ErrNotFound)
}

// An error from then leaves the history as it stood.
func TestStore_ExpireUndo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree")}
	first, _ := s.Write(ctx, "a.txt", []byte("one"), "ops")
	second, _ := s.Write(ctx, "a.txt", []byte("two"), "ops")
	err := s.Expire(ctx, second, func(artifactstore.Rewrite) error { return context.Canceled })
	assert.ErrorIs(t, err, context.Canceled)
	commits, _ := s.Commits(ctx)
	assert.Equal(t, []string{first, second}, commits)
	old, err := s.Show(ctx, first, "a.txt")
	assert.NoError(t, err)
	assert.Equal(t, "one", string(old))
}
