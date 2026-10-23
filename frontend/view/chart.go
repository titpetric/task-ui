package view

import (
	"fmt"
	"html/template"
	"math"
	"sort"
	"strings"
	"time"
)

// Chart is a month of measures drawn the way rrdtool draws them: a line
// a series over a day axis with a gridline a week, a value axis of three
// or four gridlines in the chart's unit, and a gap where a series has no
// data. SVG is the drawing, Legend each series' least, mean and most,
// Total the stack's for a stacked chart. Empty says no series has a
// point in the span.
type Chart struct {
	Title  string
	SVG    template.HTML
	Legend []ChartLegend
	Total  *ChartLegend
	Empty  bool
}

// The drawing's size in its own units, and the margins of the plot: the
// value labels on the left, the dates under it.
const (
	chartWidth  = 400.0
	chartHeight = 140.0
	chartLeft   = 52.0
	chartRight  = 6.0
	chartTop    = 8.0
	chartBottom = 20.0
)

// chartStep is the span of a record when the series have no two points
// to tell it by: the hub's eight hour records.
const chartStep = 8 * time.Hour

// NewChart draws series over the span from to to.
func NewChart(title string, unit ChartUnit, from, to time.Time, series ...ChartSeries) Chart {
	return newChart(title, unit, from, to, false, series)
}

// NewStackedChart draws each series as an area over the one before; the series share their times.
func NewStackedChart(title string, unit ChartUnit, from, to time.Time, series ...ChartSeries) Chart {
	return newChart(title, unit, from, to, true, series)
}

// newChart draws the chart.
func newChart(title string, unit ChartUnit, from, to time.Time, stacked bool, series []ChartSeries) Chart {
	c := Chart{Title: title}
	for i := range series {
		series[i].Points = within(series[i].Points, from, to)
	}
	drawn := series
	if stacked {
		drawn = stack(series)
	}
	largest := unit.Max
	points := 0
	for _, s := range drawn {
		for _, p := range s.Points {
			if !math.IsNaN(p.V) {
				largest = math.Max(largest, p.V)
				points++
			}
		}
	}
	if points == 0 {
		c.Empty = true
		return c
	}
	divisor, suffix := unit.Scale(largest)
	top, step := niceScale(largest / divisor)
	if unit.Max == 100 && largest <= 100 {
		top, step = 100, 25
	}
	top, step = top*divisor, step*divisor

	plot := chartPlot{from: from, to: to, top: top}
	var svg strings.Builder
	fmt.Fprintf(&svg, `<svg class="chart-svg" viewBox="0 0 %g %g" role="img" aria-label="%s">`, chartWidth, chartHeight, template.HTMLEscapeString(title))
	for v := 0.0; v <= top+step/1000; v += step {
		y := plot.y(v)
		fmt.Fprintf(&svg, `<line class="grid" x1="%g" x2="%g" y1="%.1f" y2="%.1f"/>`, chartLeft, chartWidth-chartRight, y, y)
		fmt.Fprintf(&svg, `<text class="tick" x="%g" y="%.1f" text-anchor="end">%s</text>`, chartLeft-4, y+3, word(v, divisor, suffix))
	}
	for day := midnight(from); !day.After(to); day = day.AddDate(0, 0, 1) {
		if day.Before(from) {
			continue
		}
		x := plot.x(day)
		if day.Weekday() == time.Monday {
			fmt.Fprintf(&svg, `<line class="grid week" x1="%.1f" x2="%.1f" y1="%g" y2="%g"/>`, x, x, chartTop, chartHeight-chartBottom)
			fmt.Fprintf(&svg, `<text class="tick" x="%.1f" y="%g" text-anchor="middle">%s</text>`, x, chartHeight-5, day.Format("Jan 2"))
		}
		fmt.Fprintf(&svg, `<line class="day" x1="%.1f" x2="%.1f" y1="%g" y2="%g"/>`, x, x, chartHeight-chartBottom, chartHeight-chartBottom+3)
	}
	limit := gapLimit(drawn)
	for i := len(drawn) - 1; i >= 0; i-- {
		s := drawn[i]
		for _, run := range runs(s.Points, limit) {
			switch {
			case stacked:
				var below []ChartPoint
				if i > 0 {
					below = pointsAt(drawn[i-1].Points, run)
				}
				fmt.Fprintf(&svg, `<path class="area s%d" d="%s"/>`, s.Color, plot.band(run, below))
			case len(series) == 1:
				fmt.Fprintf(&svg, `<path class="area s%d" d="%s"/>`, s.Color, plot.band(run, nil))
			}
			if len(run) == 1 {
				fmt.Fprintf(&svg, `<circle class="dot s%d" cx="%.1f" cy="%.1f" r="1.6"/>`, s.Color, plot.x(run[0].At), plot.y(run[0].V))
				continue
			}
			fmt.Fprintf(&svg, `<path class="line s%d" d="%s"/>`, s.Color, plot.line(run))
		}
	}
	fmt.Fprintf(&svg, `<line class="axis" x1="%g" x2="%g" y1="%g" y2="%g"/>`, chartLeft, chartWidth-chartRight, chartHeight-chartBottom, chartHeight-chartBottom)
	svg.WriteString(`</svg>`)
	c.SVG = template.HTML(svg.String())

	for _, s := range series {
		if legend, ok := legendOf(s, divisor, suffix); ok {
			c.Legend = append(c.Legend, legend)
		}
	}
	if stacked && len(series) > 1 {
		if total, ok := legendOf(ChartSeries{Name: "Total", Color: -1, Points: drawn[len(drawn)-1].Points}, divisor, suffix); ok {
			c.Total = &total
		}
	}
	return c
}

// chartPlot maps a time and a value to the drawing.
type chartPlot struct {
	from, to time.Time
	top      float64
}

func (p chartPlot) x(at time.Time) float64 {
	span := p.to.Sub(p.from).Seconds()
	return chartLeft + at.Sub(p.from).Seconds()/span*(chartWidth-chartLeft-chartRight)
}

func (p chartPlot) y(v float64) float64 {
	return chartHeight - chartBottom - v/p.top*(chartHeight-chartTop-chartBottom)
}

// line is the path through a run of points.
func (p chartPlot) line(run []ChartPoint) string {
	var d strings.Builder
	for i, point := range run {
		cmd := "L"
		if i == 0 {
			cmd = "M"
		}
		fmt.Fprintf(&d, "%s%.1f %.1f", cmd, p.x(point.At), p.y(point.V))
	}
	return d.String()
}

// band is the area between a run and the points under it, the base line
// without them.
func (p chartPlot) band(run, below []ChartPoint) string {
	d := p.line(run)
	for i := len(run) - 1; i >= 0; i-- {
		v := 0.0
		if below != nil && !math.IsNaN(below[i].V) {
			v = below[i].V
		}
		d += fmt.Sprintf("L%.1f %.1f", p.x(run[i].At), p.y(v))
	}
	return d + "Z"
}

// within keeps the points of the span.
func within(points []ChartPoint, from, to time.Time) []ChartPoint {
	var kept []ChartPoint
	for _, p := range points {
		if !p.At.Before(from) && !p.At.After(to) {
			kept = append(kept, p)
		}
	}
	return kept
}

// stack sums the series in order: each point the sum of its own and
// those of the series before at its time, a missing one counting 0.
func stack(series []ChartSeries) []ChartSeries {
	stacked := make([]ChartSeries, len(series))
	sums := map[time.Time]float64{}
	for i, s := range series {
		stacked[i] = ChartSeries{Name: s.Name, Color: s.Color}
		for _, p := range s.Points {
			v := p.V
			if math.IsNaN(v) {
				v = 0
			}
			sums[p.At] += v
			stacked[i].Points = append(stacked[i].Points, ChartPoint{At: p.At, V: sums[p.At]})
		}
	}
	return stacked
}

// pointsAt returns, for each point of a run, the point of others at its
// time, NaN where others has none.
func pointsAt(others, run []ChartPoint) []ChartPoint {
	at := map[time.Time]float64{}
	for _, p := range others {
		at[p.At] = p.V
	}
	found := make([]ChartPoint, len(run))
	for i, p := range run {
		v, ok := at[p.At]
		if !ok {
			v = math.NaN()
		}
		found[i] = ChartPoint{At: p.At, V: v}
	}
	return found
}

// gapLimit is the longest span between two points the line bridges: one
// and a half of the shortest span between two points of a series.
func gapLimit(series []ChartSeries) time.Duration {
	step := time.Duration(0)
	for _, s := range series {
		for i := 1; i < len(s.Points); i++ {
			if d := s.Points[i].At.Sub(s.Points[i-1].At); d > 0 && (step == 0 || d < step) {
				step = d
			}
		}
	}
	if step == 0 {
		step = chartStep
	}
	return step * 3 / 2
}

// runs splits points at a gap, a NaN or a span past limit.
func runs(points []ChartPoint, limit time.Duration) [][]ChartPoint {
	var all [][]ChartPoint
	var run []ChartPoint
	for i, p := range points {
		if math.IsNaN(p.V) || (i > 0 && p.At.Sub(points[i-1].At) > limit) {
			if len(run) > 0 {
				all = append(all, run)
			}
			run = nil
		}
		if !math.IsNaN(p.V) {
			run = append(run, p)
		}
	}
	if len(run) > 0 {
		all = append(all, run)
	}
	return all
}

// legendOf words a series' least, mean and most; false for a series
// without points.
func legendOf(s ChartSeries, divisor float64, suffix string) (ChartLegend, bool) {
	var values []float64
	for _, p := range s.Points {
		if !math.IsNaN(p.V) {
			values = append(values, p.V)
		}
	}
	if len(values) == 0 {
		return ChartLegend{}, false
	}
	sort.Float64s(values)
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return ChartLegend{
		Name: s.Name, Color: s.Color,
		Min: word(values[0], divisor, suffix), Avg: word(sum/float64(len(values)), divisor, suffix), Max: word(values[len(values)-1], divisor, suffix),
	}, true
}

// niceScale returns a top at or over largest and a step that divides it
// in three or four: 1, 2, 2.5 or 5 times a power of ten.
func niceScale(largest float64) (top, step float64) {
	if largest <= 0 {
		return 1, 0.5
	}
	raw := largest / 3
	magnitude := math.Pow(10, math.Floor(math.Log10(raw)))
	for _, m := range []float64{1, 2, 2.5, 5, 10} {
		if step = m * magnitude; step >= raw {
			break
		}
	}
	return step * math.Ceil(largest/step-1e-9), step
}

// midnight is the start of the day of a time, in its zone.
func midnight(at time.Time) time.Time {
	y, m, d := at.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, at.Location())
}
