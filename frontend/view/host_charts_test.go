package view_test

import (
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/hostinfo"
	"github.com/titpetric/task-ui/internal/assert"
)

// samples are three records eight hours apart, the last now.
func samples(stats func(i int) beszel.Stats) []beszel.Sample {
	var s []beszel.Sample
	for i := 0; i < 3; i++ {
		s = append(s, beszel.Sample{At: chartNow.Add(-time.Duration(2-i) * 8 * time.Hour), Stats: stats(i)})
	}
	return s
}

// A host's month: CPU, memory, a disk line each, the traffic both ways.
func TestHostCharts(t *testing.T) {
	t.Parallel()
	month := samples(func(i int) beszel.Stats {
		return beszel.Stats{CPU: 10, Memory: 8, MemoryUsed: 4, Disk: 100, DiskUsed: 50, Bandwidth: beszel.Pair{100, 200}, ExtraFs: map[string]beszel.FsStats{"data": {Total: 10, Used: float64(i)}}}
	})
	charts := view.HostCharts(beszel.System{}, month, chartNow)
	assert.Len(t, charts, 4)
	assert.Equal(t, "CPU", charts[0].Title)
	assert.Equal(t, "4 GiB", charts[1].Legend[0].Avg)
	assert.Equal(t, []string{"/", "data"}, []string{charts[2].Legend[0].Name, charts[2].Legend[1].Name})
	assert.Equal(t, "50%", charts[2].Legend[0].Max)
	assert.Equal(t, "200 B/s", charts[3].Legend[1].Max)
}

// The containers' month: the five busiest, the busiest first.
func TestContainerCharts(t *testing.T) {
	t.Parallel()
	var month []beszel.ContainerSample
	for i := 0; i < 3; i++ {
		s := beszel.ContainerSample{At: chartNow.Add(-time.Duration(i) * 8 * time.Hour)}
		for n, name := range []string{"a", "b", "c", "d", "e", "f"} {
			s.Containers = append(s.Containers, beszel.Container{Name: name, Memory: float64(n * 100), CPU: float64(6 - n)})
		}
		month = append(month, s)
	}
	charts := view.ContainerCharts(chartNow, month)
	assert.Len(t, charts, 2)
	assert.Len(t, charts[0].Legend, view.ChartTop)
	assert.Equal(t, "f", charts[0].Legend[0].Name)
	assert.Equal(t, "a", charts[1].Legend[0].Name)
	assert.Equal(t, 0, charts[1].Legend[0].Color)
}

// The cluster's month stacks the hosts of the pools in their colours,
// leaving out the ones the pools leave out.
func TestClusterCharts(t *testing.T) {
	t.Parallel()
	systems := []beszel.System{
		{ID: "a", Name: "big", Status: "up", Info: beszel.Info{Threads: 8}, Stats: &beszel.Stats{Memory: 16}},
		{ID: "b", Name: "small", Status: "up", Info: beszel.Info{Threads: 2}, Stats: &beszel.Stats{Memory: 4}},
		{ID: "c", Name: "gone", Status: "down"},
	}
	cluster := hostinfo.NewCluster(systems, nil, nil)
	history := map[string][]beszel.Sample{
		"a": samples(func(int) beszel.Stats { return beszel.Stats{CPU: 50, MemoryUsed: 8} }),
		"b": samples(func(int) beszel.Stats { return beszel.Stats{CPU: 100, MemoryUsed: 2} })[1:],
		"c": samples(func(int) beszel.Stats { return beszel.Stats{CPU: 100} }),
	}
	charts := view.ClusterCharts(&cluster, systems, history, chartNow)
	assert.Len(t, charts, 2)
	cores := charts[0]
	assert.Len(t, cores.Legend, 2)
	assert.Equal(t, view.ChartLegend{Name: "small", Color: 1, Min: "2 cores", Avg: "2 cores", Max: "2 cores"}, cores.Legend[1])
	assert.Equal(t, "4 cores", cores.Total.Min)
	assert.Equal(t, "6 cores", cores.Total.Max)
	assert.True(t, strings.Contains(string(cores.SVG), `class="area s1"`))
	assert.Equal(t, "10 GiB", charts[1].Total.Max)
}
