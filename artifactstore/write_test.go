package artifactstore_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/internal/assert"
)

// Write puts one file into the tree over the version there, which the
// history keeps; it refuses a path out of the tree, a dot file, and a
// file or a directory in the way.
func TestStore_Write(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree")}

	first, err := s.Write(ctx, "notes/plan.md", []byte("one"), "ops")
	assert.NoError(t, err)
	second, err := s.Write(ctx, "notes/plan.md", []byte("two"), "ops")
	assert.NoError(t, err)
	assert.NotEqual(t, first, second)
	assert.Equal(t, "two", read(t, s, "notes/plan.md"))
	old, err := s.Show(ctx, first, "notes/plan.md")
	assert.NoError(t, err)
	assert.Equal(t, "one", string(old))
	same, err := s.Write(ctx, "notes/plan.md", []byte("two"), "ops")
	assert.NoError(t, err)
	assert.Equal(t, second, same)

	for _, rel := range []string{"", "/abs", "a/../b", "..", ".env", "a/.git/x", "a//b", "a/", `a\b`} {
		_, err := s.Write(ctx, rel, []byte("x"), "ops")
		assert.ErrorIs(t, err, artifactstore.ErrPath, rel)
	}
	_, err = s.Write(ctx, "notes", []byte("x"), "ops")
	assert.ErrorIs(t, err, artifactstore.ErrInTheWay)
	_, err = s.Write(ctx, "notes/plan.md/inner", []byte("x"), "ops")
	assert.ErrorIs(t, err, artifactstore.ErrInTheWay)
}

func TestStore_WriteThen(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree"), Keep: 1}
	var got string
	commit, err := s.WriteThen(ctx, "a.txt", []byte("a"), "ops", func(c string) error { got = c; return nil })
	assert.NoError(t, err)
	assert.Equal(t, commit, got)
}
