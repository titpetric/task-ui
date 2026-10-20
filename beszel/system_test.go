package beszel_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestSystem(t *testing.T) {
	t.Parallel()
	s := beszel.System{Status: "paused", Info: beszel.Info{Uptime: 90}}
	assert.False(t, s.Up())
	assert.Equal(t, 90*time.Second, s.UptimeDuration())
}
