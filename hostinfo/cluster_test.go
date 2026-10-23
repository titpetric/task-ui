package hostinfo_test

import (
	"testing"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/hostinfo"
	"github.com/titpetric/task-ui/internal/assert"
)

// The pools sum the hosts that are up and report, a segment a host as
// wide as its share; a host down or without measures counts only as a
// host, greyed in the legend.
func TestNewCluster(t *testing.T) {
	t.Parallel()
	systems := []beszel.System{
		{ID: "a", Name: "big", Status: "up", Info: beszel.Info{Threads: 12, CPU: 50}, Stats: &beszel.Stats{Memory: 48, MemoryUsed: 12, Disk: 100, DiskUsed: 50}},
		{ID: "b", Name: "small", Status: "up", Info: beszel.Info{Threads: 4, CPU: 25}, Stats: &beszel.Stats{Memory: 16, MemoryUsed: 8, Disk: 100, DiskUsed: 10}},
		{ID: "c", Name: "gone", Status: "down", Info: beszel.Info{Threads: 64}, Stats: &beszel.Stats{Memory: 512}},
		{ID: "d", Name: "mute", Status: "up", Info: beszel.Info{Threads: 8}},
	}
	containers := map[string][]beszel.Container{"a": {{Name: "x"}, {Name: "y"}, {Name: "z"}}, "b": {{Name: "w"}}, "c": {{Name: "v"}}}
	disks := map[string][]hostinfo.Filesystem{"b": {{Size: 300 << 30, Used: 100 << 30}, {Size: 100 << 30, Used: 0}}}
	c := hostinfo.NewCluster(systems, containers, disks)

	assert.Equal(t, 3, c.Up)
	assert.Equal(t, 1, c.Down)
	assert.Equal(t, 2, c.Left)
	assert.Equal(t, []hostinfo.ClusterHost{{Name: "big", Color: 0}, {Name: "small", Color: 1}, {Name: "gone", Color: -1, Left: true}, {Name: "mute", Color: -1, Left: true}}, c.Hosts)

	cpu := c.Pools[0]
	assert.Equal(t, "CPU", cpu.Name)
	assert.Equal(t, 16.0, cpu.Total)
	assert.Equal(t, 7.0, cpu.Used)
	assert.Equal(t, "7 / 16 cores", cpu.Figure)
	assert.Equal(t, 43.75, cpu.Percent())
	assert.Equal(t, 75.0, cpu.Segments[0].Width)
	assert.Equal(t, 50.0, cpu.Segments[0].Fill)
	assert.Equal(t, "big: 6 cores of 12 cores", cpu.Segments[0].Title)

	memory := c.Pools[1]
	assert.Equal(t, "20.0 / 64.0 GiB", memory.Figure)
	assert.Equal(t, 1, memory.Segments[1].Color)

	disk := c.Pools[2]
	assert.Equal(t, 500.0, disk.Total, "the durable filesystems of small, Beszel's root of big")
	assert.Equal(t, 150.0, disk.Used)
	assert.Equal(t, "small: 100.0 GiB of 400.0 GiB", disk.Segments[1].Title)

	running := c.Pools[3]
	assert.Equal(t, "4 running", running.Figure)
	assert.Equal(t, 75.0, running.Segments[0].Width)
	assert.Equal(t, 100.0, running.Segments[0].Fill)
	assert.Equal(t, "big: 3 running", running.Segments[0].Title)
	assert.Equal(t, 0.0, hostinfo.Pool{}.Percent())
}

// A cluster of terabytes says so.
func TestNewCluster_TiB(t *testing.T) {
	t.Parallel()
	systems := []beszel.System{{ID: "a", Name: "nas", Status: "up", Stats: &beszel.Stats{Disk: 4096, DiskUsed: 1024, Memory: 1}}}
	c := hostinfo.NewCluster(systems, nil, nil)
	assert.Equal(t, "1.0 / 4.0 TiB", c.Pools[2].Figure)
	assert.Equal(t, "nas: 1.0 TiB of 4.0 TiB", c.Pools[2].Segments[0].Title)
	assert.Equal(t, hostinfo.Segment{Host: "nas", Used: 0, Total: 0, Title: "nas: 0 cores of 0 cores"}, c.Pools[0].Segments[0])
}

// A cluster of no hosts has empty pools.
func TestCluster(t *testing.T) {
	t.Parallel()
	c := hostinfo.NewCluster(nil, nil, nil)
	assert.Equal(t, hostinfo.Cluster{Pools: c.Pools}, c)
	assert.Len(t, c.Pools, 4)
	assert.Equal(t, "0 / 0 cores", c.Pools[0].Figure)
}
