//go:build integration

package artifactstore_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/internal/assert"
)

// The history of a folder is its own commits, newest first, with the
// paths each one changed. It is the half of a rollback that says what to
// go back to; Show is the half that reads it.
func TestStore_History(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree")}

	assert.Empty(t, mustHistory(t, s, "", 10), "a store with no repository has no history")

	_, err := s.Write(ctx, "01THREAD/diagrams/schema.svg", []byte("<svg>one</svg>"), "claude")
	assert.NoError(t, err)
	_, err = s.Write(ctx, "01OTHER/notes.md", []byte("not this thread\n"), "claude")
	assert.NoError(t, err)
	_, err = s.Write(ctx, "01THREAD/diagrams/schema.svg", []byte("<svg>two</svg>"), "claude")
	assert.NoError(t, err)

	history := mustHistory(t, s, "01THREAD", 10)
	assert.Len(t, history, 2, "the other thread's commit is not this folder's history")
	assert.Equal(t, []string{"01THREAD/diagrams/schema.svg"}, history[0].Files)
	assert.Contains(t, history[0].Subject, "schema.svg")
	assert.Equal(t, 40, len(history[0].Commit))
	assert.Equal(t, 7, len(history[0].Short))
	assert.NotEqual(t, "", history[0].When)

	// Newest first, and the older one is what a rollback reads.
	assert.NotEqual(t, history[0].Commit, history[1].Commit)
	was, err := s.Show(ctx, history[1].Commit, "01THREAD/diagrams/schema.svg")
	assert.NoError(t, err)
	assert.Equal(t, "<svg>one</svg>", string(was))

	assert.Len(t, mustHistory(t, s, "01THREAD", 1), 1, "the limit bounds what comes back")
	assert.Empty(t, mustHistory(t, s, "01THREAD", 0), "no limit is no history")
	assert.Empty(t, mustHistory(t, s, "01NOTHING", 10), "a folder no commit touched")

	_, err = s.History(ctx, "../escape", 10)
	assert.Error(t, err, "a path that leaves the tree has no history")
}

func mustHistory(t *testing.T, s *artifactstore.Store, rel string, limit int) []artifactstore.Revision {
	t.Helper()
	history, err := s.History(context.Background(), rel, limit)
	assert.NoError(t, err)
	return history
}
