package api_test

import (
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

// terminal is a run in progress that keeps what windows send it.
type terminal struct {
	input   strings.Builder
	taken   []string
	stopped bool
}

func (t *terminal) Input(p []byte) error {
	t.input.Write(p)
	return nil
}

func (t *terminal) TakeInput(user string) bool {
	t.taken = append(t.taken, user)
	return len(t.taken) == 1
}

func (t *terminal) Resize(int, int) {}

func (t *terminal) Stop(time.Duration) { t.stopped = true }

func (t *terminal) Kill(time.Duration) { t.stopped = true }

func TestRunTerminal(t *testing.T) {
	t.Parallel()
	// A session is what a window drives.
	var _ api.RunTerminal = (*runner.Session)(nil)
	var fake api.RunTerminal = &terminal{}
	assert.True(t, fake.TakeInput("black"))
	assert.False(t, fake.TakeInput("black"))
}
