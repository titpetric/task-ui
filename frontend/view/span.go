package view

import (
	"sort"
	"time"
)

// Span is the window a chart covers: how many slots it has and how
// wide each one is, counted in the server's time zone. Activity and
// Tooling chart the same spans, so their columns line up.
//
// Its ticks mark the month a day a week, and the last day; the day every third clock hour, labelled
// at 03:00, 09:00, 15:00 and 21:00; and the week each day at its noon.
type Span struct {
	// Title heads the chart, and Unit names one slot: "busiest day".
	Title string
	Unit  string

	// Slots is how many columns the chart draws, and Hours how wide
	// each is; zero is a calendar day.
	Slots int
	Hours int

	// Wide takes a row of its own: a week in three-hour slots needs
	// the width.
	Wide bool
}

// The spans the runs page charts.
var (
	SpanMonth = Span{Title: "Last 30 days", Unit: "day", Slots: 30}
	SpanDay   = Span{Title: "Last 24 hours", Unit: "hour", Slots: 24, Hours: 1}
	SpanWeek  = Span{Title: "Last 7 days", Unit: "3 hours", Slots: 56, Hours: 3, Wide: true}
)

// Spans are the charts in the order the page draws them: the month and the
// day side by side, the week under them on a row of its own.
var Spans = []Span{SpanMonth, SpanDay, SpanWeek}

// Days is how many calendar days the span reaches back.
func (s Span) Days() int {
	if s.Hours == 0 {
		return s.Slots
	}
	return max(1, s.Slots*s.Hours/24)
}

// Edges returns where the slots start, oldest first, and where the last
// one ends: Slots+1 times, the slot now is in the last.
func (s Span) Edges(now time.Time) []time.Time {
	edges := make([]time.Time, s.Slots+1)
	y, m, d := now.Date()
	loc := now.Location()
	if s.Hours == 0 {
		first := time.Date(y, m, d, 0, 0, 0, 0, loc).AddDate(0, 0, 1-s.Slots)
		for i := range edges {
			edges[i] = first.AddDate(0, 0, i)
		}
		return edges
	}
	last := now.Hour() / s.Hours * s.Hours
	for i := range edges {
		edges[i] = time.Date(y, m, d, last+s.Hours*(i-s.Slots+1), 0, 0, 0, loc)
	}
	return edges
}

// Index returns the slot a time falls in, -1 outside the span.
func (s Span) Index(at time.Time, edges []time.Time) int {
	if len(edges) < 2 || at.Before(edges[0]) || !at.Before(edges[len(edges)-1]) {
		return -1
	}
	return sort.Search(len(edges), func(i int) bool { return edges[i].After(at) }) - 1
}

// Label names a slot for its hover: the day, or the day and the hours.
func (s Span) Label(edges []time.Time, i int) string {
	start := edges[i]
	if s.Hours == 0 {
		return start.Format("Mon 2006-01-02")
	}
	return start.Format("Mon ") + twoDigits(start.Hour()) + ":00–" + twoDigits(edges[i+1].Hour()) + ":00"
}

// Ticks mark the span under its chart, as Span describes.
func (s Span) Ticks(edges []time.Time) []ActivityTick {
	n := s.Slots
	middle := func(i int) string { return svgNumber((float64(i) + 0.5) * 100 / float64(n)) }
	var ticks []ActivityTick
	switch {
	case s.Hours == 0:
		for i := n - 1; i >= 0; i -= 7 {
			ticks = append([]ActivityTick{{Label: edges[i].Format("Jan 2"), At: middle(i)}}, ticks...)
		}
	case s.Hours == 1:
		for i := range n {
			clock := edges[i].Hour()
			if clock%3 != 0 {
				continue
			}
			tick := ActivityTick{At: middle(i)}
			if clock%6 == 3 {
				tick.Label = twoDigits(clock) + ":00"
			}
			ticks = append(ticks, tick)
		}
	default:
		for i := 1; i < n; i++ {
			if edges[i].Hour() == 12 {
				ticks = append(ticks, ActivityTick{Label: edges[i].Format("Mon"), At: svgNumber(float64(i) * 100 / float64(n))})
			}
		}
	}
	return ticks
}

// tipEdge anchors the hover of slot i of n near either end of its chart,
// so it stays inside it: tip-start, tip-end or nothing.
func tipEdge(i, n int) string {
	k := max(4, n*2/15)
	switch {
	case i < k:
		return "tip-start"
	case i >= n-k:
		return "tip-end"
	}
	return ""
}
