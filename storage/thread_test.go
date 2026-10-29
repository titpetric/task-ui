package storage_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// A thread with no reply has no last time, and the cleanup reads an
// open one as busy at any age.
func TestThread(t *testing.T) {
	t.Parallel()
	var none storage.Thread
	assert.Nil(t, none.LastAt)
	assert.False(t, none.Open)
	at := time.Now()
	thread := storage.Thread{ID: "n1", LastAt: &at, Open: true}
	assert.Equal(t, at, *thread.LastAt)
	assert.True(t, thread.Open)
}
