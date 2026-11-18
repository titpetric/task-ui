package storage_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// An empty filter matches every thread and a live one narrows to running work.
func TestConversationFilter(t *testing.T) {
	assert.False(t, storage.ConversationFilter{}.Live)
	filter := storage.ConversationFilter{WorkspaceID: "ws", Folder: "ws/app", Live: true, Limit: 10}
	assert.Equal(t, "ws/app", filter.Folder)
	assert.True(t, filter.Live)
}
