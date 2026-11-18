package claudesession_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/claudesession"
)

// A name keeps the user readable, refuses a path, and holds the session ID after the last underscore.
func TestName(t *testing.T) {
	t.Parallel()
	id := "01arz3ndektsv4rrffq69g5fav"
	assert.Equal(t, "ana_"+id, claudesession.Name("ana", id))
	assert.Equal(t, "ana_bob_"+id, claudesession.Name("ana/bob", id))
	assert.Equal(t, "user_"+id, claudesession.Name("..", id))
	got, ok := claudesession.ID(claudesession.Name("ana_bob", id))
	assert.True(t, ok)
	assert.Equal(t, id, got)
	_, ok = claudesession.ID("no-session")
	assert.False(t, ok)
}

// A directory is the root, the user and the session.
func TestDir(t *testing.T) {
	t.Parallel()
	id := "01arz3ndektsv4rrffq69g5fav"
	assert.Equal(t, filepath.Join("/data/claude", "ana_"+id), claudesession.Dir("/data/claude", "ana", id))
	assert.Equal(t, "", claudesession.Dir("", "ana", id))
	assert.Equal(t, "", claudesession.Dir("/data/claude", "ana", ""))
}

// A session maps to one UUID, which Claude Code takes where the ID is a ULID.
func TestUUID(t *testing.T) {
	t.Parallel()
	id := "01arz3ndektsv4rrffq69g5fav"
	first := claudesession.UUID(id)
	_, err := uuid.Parse(first)
	assert.NoError(t, err)
	assert.Equal(t, first, claudesession.UUID(id))
	assert.NotEqual(t, first, claudesession.UUID("01arz3ndektsv4rrffq69g5faw"))
}

// A remove takes the directory, and an empty path is a no-op.
func TestRemove(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir := claudesession.Dir(root, "ana", "01arz3ndektsv4rrffq69g5fav")
	assert.NoError(t, os.MkdirAll(dir, 0o700))
	assert.NoError(t, claudesession.Remove(dir))
	_, err := os.Stat(dir)
	assert.True(t, os.IsNotExist(err))
	assert.NoError(t, claudesession.Remove(""))
}

// A clean keeps the sessions it names and removes the others.
func TestClean(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	keep, drop := "01arz3ndektsv4rrffq69g5fav", "01arz3ndektsv4rrffq69g5faw"
	for _, id := range []string{keep, drop} {
		assert.NoError(t, os.MkdirAll(claudesession.Dir(root, "ana", id), 0o700))
	}
	assert.NoError(t, os.WriteFile(filepath.Join(root, "not-a-session"), []byte("x"), 0o600))
	assert.NoError(t, claudesession.Clean(root, map[string]bool{keep: true}))
	_, err := os.Stat(claudesession.Dir(root, "ana", drop))
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(claudesession.Dir(root, "ana", keep))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(root, "not-a-session"))
	assert.NoError(t, err)
}
