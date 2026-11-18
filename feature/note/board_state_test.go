package note

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

// A shutdown's kill leaves the run's session for Recover; a user's stop removes it.
func TestBoard_Close(t *testing.T) {
	t.Parallel()
	board := &Board{ClaudeDir: t.TempDir()}
	assert.False(t, board.stopped(runner.Request{}))
	board.Close()
	assert.True(t, board.stopped(runner.Request{}))
}
