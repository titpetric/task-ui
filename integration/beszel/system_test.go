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

// Stats carry the sizes the systems list leaves out.
func TestStats(t *testing.T) {
	t.Parallel()
	stats := &beszel.Stats{Memory: 7.7, MemoryUsed: 6.2, Disk: 196, DiskUsed: 82}
	assert.Equal(t, 7.7, beszel.System{Stats: stats}.Stats.Memory)
}

// The largest hosts come first: cores, then RAM, then the name.
func TestByCapacity(t *testing.T) {
	t.Parallel()
	gb := func(m float64) *beszel.Stats { return &beszel.Stats{Memory: m} }
	systems := []beszel.System{
		{Name: "mgmt", Info: beszel.Info{Threads: 2}, Stats: gb(3.8)},
		{Name: "docker1", Info: beszel.Info{Threads: 4}, Stats: gb(7.7)},
		{Name: "pve", Info: beszel.Info{Threads: 16}, Stats: gb(62.7)},
		{Name: "big", Info: beszel.Info{Threads: 4}, Stats: gb(15.6)},
		{Name: "dev1", Info: beszel.Info{Threads: 4}, Stats: gb(7.7)},
		{Name: "bare", Info: beszel.Info{Threads: 4}},
	}
	names := []string{}
	for _, s := range beszel.ByCapacity(systems) {
		names = append(names, s.Name)
	}
	assert.Equal(t, []string{"pve", "big", "dev1", "docker1", "bare", "mgmt"}, names)
	assert.Equal(t, "mgmt", systems[0].Name, "the input stays as it was")
}
