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
	assert.Equal(t, 0, s.CPUs())
	assert.Equal(t, 4, beszel.System{Info: beszel.Info{Threads: 4}}.CPUs())
	assert.Equal(t, 8, beszel.System{Info: beszel.Info{Cores: 8, Threads: 16}}.CPUs())
}
