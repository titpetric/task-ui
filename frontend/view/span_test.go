package view_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/tooling"
)

// A span's slots end with the slot now is in, counted on the clock of
// now's zone: the week's three-hour slots start at 00:00, 03:00 and so
// on there.
func TestSpan(t *testing.T) {
	t.Parallel()
	now := at(19, 22, 10) // 20:10 UTC

	week := view.SpanWeek.Edges(now)
	assert.Len(t, week, 57)
	assert.True(t, week[55].Equal(at(19, 21, 0)))
	assert.True(t, week[56].Equal(at(20, 0, 0)))
	assert.True(t, week[0].Equal(at(12, 24, 0)))
	assert.Equal(t, 55, view.SpanWeek.Index(now, week))
	assert.Equal(t, 0, view.SpanWeek.Index(at(13, 0, 0), week))
	assert.Equal(t, -1, view.SpanWeek.Index(at(12, 23, 59), week))
	assert.Equal(t, -1, view.SpanWeek.Index(at(20, 0, 0), week))
	assert.Equal(t, "Mon 21:00–00:00", view.SpanWeek.Label(week, 55))
	assert.Equal(t, 7, view.SpanWeek.Days())

	// The week's ticks name each day at its noon.
	ticks := view.SpanWeek.Ticks(week)
	assert.Len(t, ticks, 7)
	assert.Equal(t, view.ActivityTick{Label: "Mon", At: "92.86"}, ticks[6])

	day := view.SpanDay.Edges(now)
	assert.Len(t, day, 25)
	assert.True(t, day[23].Equal(at(19, 22, 0)))
	assert.Equal(t, "Mon 22:00–23:00", view.SpanDay.Label(day, 23))
	assert.Equal(t, 1, view.SpanDay.Days())

	month := view.SpanMonth.Edges(now)
	assert.Len(t, month, 31)
	assert.True(t, month[29].Equal(at(19, 0, 0)))
	assert.Equal(t, "Mon 2026-10-19", view.SpanMonth.Label(month, 29))
	assert.Equal(t, 30, view.SpanMonth.Days())
	assert.Equal(t, 29, view.SpanMonth.Index(at(19, 1, 0), month)) // the 18th in UTC

	assert.Len(t, view.Spans, 3)
	assert.True(t, view.Spans[2].Wide)
}

// With samples the panel draws the activity charts' spans: the month by
// day, the day by hour and the week by three hours, counted in now's
// zone, each with its own top five and other.
func TestToolingFromSamples(t *testing.T) {
	t.Parallel()

	scope := view.NewTooling(&tooling.Report{Project: sampleScope()}, toolingNow).Scopes[0]
	assert.Len(t, scope.Windows, 3)

	month, day, week := scope.Windows[0], scope.Windows[1], scope.Windows[2]
	assert.Equal(t, "Last 30 days", month.Title)
	assert.Len(t, month.Bars, 30)
	assert.Equal(t, "Last 24 hours", day.Title)
	assert.Len(t, day.Bars, 24)
	assert.Equal(t, "hour", day.Unit)
	assert.Equal(t, "Last 7 days", week.Title)
	assert.Len(t, week.Bars, 56)
	assert.True(t, week.Wide)
	assert.Equal(t, "fine", week.Density)

	// The month holds the fortnight-old build, the day and the week do
	// not: docker build leads the month and is under go test on the day.
	assert.Equal(t, int64(32), month.Runs)
	assert.Equal(t, "docker build", month.Legend[0].Name)
	assert.Equal(t, int64(31), day.Runs)
	assert.Equal(t, "go test", day.Legend[0].Name)
	assert.Len(t, day.Legend, 6)
	assert.Equal(t, "other", day.Legend[5].Name)
	assert.Equal(t, int64(3), day.Legend[5].Runs)

	// An hour nothing ran in is an empty slot that says so; an hour
	// that ran says what.
	assert.Len(t, day.Bars[23].Segments, 6)
	assert.Contains(t, day.Bars[23].Title, "Fri 18:00–19:00: 1 command in 1.0s")
	assert.Contains(t, day.Bars[22].Title, "Fri 17:00–18:00: 8 commands in ")
	assert.Equal(t, "tip-end", day.Bars[23].Edge)

	// The week's last slot is 18:00 to 21:00 and holds the 18:00 test.
	assert.Equal(t, "Fri 18:00–21:00", week.Bars[55].Date)
	assert.Equal(t, "Fri 18:00–21:00 · go test: 1.0s, 1 run", week.Bars[55].Segments[0].Title)
	assert.Len(t, week.Bars[0].Segments, 0)

	// The week in another zone moves its slots with the clock.
	east := view.NewTooling(&tooling.Report{Project: sampleScope()}, toolingNow.In(time.FixedZone("CEST", 2*60*60))).Scopes[0]
	assert.Equal(t, "Fri 18:00–21:00", east.Windows[2].Bars[55].Date)
	assert.Equal(t, "Fri 20:00–21:00", east.Windows[1].Bars[23].Date)
}

// A report too short for the month still draws the day and the week.
func TestToolingFromSamplesShortReport(t *testing.T) {
	t.Parallel()

	report := sampleScope()
	report.Days = 7
	windows := view.NewTooling(&tooling.Report{Project: report}, toolingNow).Scopes[0].Windows
	assert.Len(t, windows, 2)
	assert.Equal(t, "Last 24 hours", windows[0].Title)
	assert.Equal(t, "Last 7 days", windows[1].Title)
}
