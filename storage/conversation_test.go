package storage_test

import (
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// A conversation is a note with its newest activity and reply count.
func TestConversation(t *testing.T) {
	row := storage.Conversation{Note: model.Note{ID: "n1", State: model.NoteWorking}, LastAt: "2026-11-02 10:00:00", Messages: 2}
	assert.Equal(t, "n1", row.ID)
	assert.Equal(t, 2, row.Messages)
	assert.Equal(t, "2026-11-02 10:00:00", row.LastAt)
}
