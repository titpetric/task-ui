package storage_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// TestParticipant keeps the fields a chat participant carries together.
func TestParticipant(t *testing.T) {
	t.Parallel()
	at := time.Now()
	participant := storage.Participant{UserID: "u1", UserName: "alice", LastReadAt: &at}
	assert.Equal(t, "alice", participant.UserName)
	assert.Equal(t, &at, participant.LastReadAt)
}
