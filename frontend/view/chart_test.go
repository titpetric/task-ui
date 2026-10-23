package view_test

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

var chartNow = time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)

// points are a series every eight hours of the month, by a function of
// the index; NaN leaves a point out.
func points(n int, v func(i int) float64) []view.ChartPoint {
	var ps []view.ChartPoint
	for i := 0; i < n; i++ {
		if value := v(i); !math.IsNaN(value) {
			ps = append(ps, view.ChartPoint{At: chartNow.Add(-time.Duration(n-1-i) * 8 * time.Hour), V: value})
		}
	}
	return ps
}

// A line breaks where its records stop, and draws a lone point as a
// dot; a percent runs from 0 to 100 in quarters, the legend under it.
func TestNewChart(t *testing.T) {
	t.Parallel()
	from := chartNow.Add(-30 * 24 * time.Hour)
	cpu := view.ChartSeries{Name: "CPU", Points: points(90, func(i int) float64 {
		if i >= 40 && i < 46 || i == 60 || i == 62 {
			return math.NaN()
		}
		return float64(i % 50)
	})}
	c := view.NewChart("CPU <1>", view.ChartPercent, from, chartNow, cpu)
	svg := string(c.SVG)
	assert.False(t, c.Empty)
	assert.Contains(t, svg, `aria-label="CPU &lt;1&gt;"`)
	assert.Equal(t, 3, strings.Count(svg, `class="line s0"`), "three runs: the gap and the lone point break it")
	assert.Equal(t, 1, strings.Count(svg, `class="dot s0"`))
	assert.Equal(t, 4, strings.Count(svg, `class="area s0"`), "a lone series has an area under each run, its dot too")
	for _, tick := range []string{">0%<", ">25%<", ">50%<", ">75%<", ">100%<"} {
		assert.Contains(t, svg, tick)
	}
	assert.Contains(t, svg, ">Oct 19<", "a Monday")
	assert.NotContains(t, svg, ">Oct 20<")
	assert.Equal(t, []view.ChartLegend{{Name: "CPU", Min: "0%", Avg: "21%", Max: "49%"}}, c.Legend)
	assert.Nil(t, c.Total)

	empty := view.NewChart("Disk", view.ChartPercent, from, chartNow, view.ChartSeries{Name: "/"}, view.ChartSeries{Points: []view.ChartPoint{{At: from.Add(-time.Hour), V: 1}}})
	assert.Equal(t, view.Chart{Title: "Disk", Empty: true}, empty)
}

// A scale fits the data in its unit: three or four gridlines, the unit
// stepping up past 1024.
func TestNewChart_Scale(t *testing.T) {
	t.Parallel()
	from := chartNow.Add(-30 * 24 * time.Hour)
	rate := view.NewChart("Net", view.ChartRate, from, chartNow, view.ChartSeries{Name: "Sent", Color: 1, Points: points(3, func(i int) float64 { return float64(i) * 1.5 * 1024 * 1024 })})
	svg := string(rate.SVG)
	assert.Contains(t, svg, ">3 MB/s<")
	assert.Contains(t, svg, ">1 MB/s<")
	assert.Contains(t, svg, `class="line s1"`)
	assert.Contains(t, svg, `class="area s1"`, "an area under a lone series")
	two := view.NewChart("Net", view.ChartRate, from, chartNow, view.ChartSeries{Points: points(2, func(int) float64 { return 1 })}, view.ChartSeries{Color: 1, Points: points(2, func(int) float64 { return 2 })})
	assert.NotContains(t, string(two.SVG), `class="area`, "no area under one of two series")
	assert.Equal(t, "1.5 MB/s", rate.Legend[0].Avg)

	cores := view.NewChart("Busy", view.ChartCores, from, chartNow, view.ChartSeries{Points: points(2, func(int) float64 { return 0 })})
	assert.Contains(t, string(cores.SVG), ">0.5 cores<")

	gib := view.NewChart("Memory", view.ChartUnit{Max: 7.7, Scale: view.ChartGiB.Scale}, from, chartNow, view.ChartSeries{Points: points(2, func(int) float64 { return 2 })})
	assert.Contains(t, string(gib.SVG), ">5 GiB<")
	assert.Contains(t, string(gib.SVG), ">10 GiB<")
	tib := view.NewChart("Disk", view.ChartGiB, from, chartNow, view.ChartSeries{Points: points(2, func(int) float64 { return 3000 })})
	assert.Contains(t, string(tib.SVG), " TiB<")
	mb := view.NewChart("Mem", view.ChartMB, from, chartNow, view.ChartSeries{Points: points(2, func(int) float64 { return 300 })})
	assert.Contains(t, string(mb.SVG), ">300 MB<")
}

// A stack draws each series over the one before; its total sums them,
// a missing point counting 0.
func TestNewStackedChart(t *testing.T) {
	t.Parallel()
	from := chartNow.Add(-30 * 24 * time.Hour)
	a := view.ChartSeries{Name: "a", Points: points(3, func(int) float64 { return 2 })}
	b := view.ChartSeries{Name: "b", Color: 1, Points: points(3, func(i int) float64 {
		if i == 1 {
			return math.NaN()
		}
		return 1
	})}
	b.Points = append(b.Points[:1], append([]view.ChartPoint{{At: a.Points[1].At, V: math.NaN()}}, b.Points[1:]...)...)
	c := view.NewStackedChart("Busy", view.ChartCores, from, chartNow, a, b)
	assert.Contains(t, string(c.SVG), `class="area s1"`)
	assert.Equal(t, "1 cores", c.Legend[1].Avg)
	assert.NotNil(t, c.Total)
	assert.Equal(t, view.ChartLegend{Name: "Total", Color: -1, Min: "2 cores", Avg: "2.7 cores", Max: "3 cores"}, *c.Total)
}

// The units word their values.
func TestChartUnit(t *testing.T) {
	t.Parallel()
	divisor, suffix := view.ChartRate.Scale(5 << 30)
	assert.Equal(t, float64(1<<30), divisor)
	assert.Equal(t, " GB/s", suffix)
	divisor, suffix = view.ChartPercent.Scale(1e9)
	assert.Equal(t, 1.0, divisor)
	assert.Equal(t, "%", suffix)
}
