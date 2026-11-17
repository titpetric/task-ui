package storage_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// TestChatSummary keeps the last message and the unread count a list draws.
func TestChatSummary(t *testing.T) {
	t.Parallel()
	summary := storage.ChatSummary{ID: "c1", Peer: "alice", LastBody: "hello", Unread: 2}
	assert.Equal(t, "hello", summary.LastBody)
	assert.Equal(t, int64(2), summary.Unread)
}
