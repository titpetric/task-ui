package view_test

import (
	"html/template"
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A zone east of UTC, so a run early in the morning is on the day before
// in UTC but on its own day here.
var ljubljana = time.FixedZone("CEST", 2*60*60)

func at(day, hour, minute int) time.Time {
	return time.Date(2026, time.October, day, hour, minute, 0, 0, ljubljana)
}

// The days are the last thirty, today the last, counted in the zone of
// now; the hours are the hours of the day over the same days. A run is
// ok, failed (killed too) or other (running, stopped); the chart leaves out
// runs outside the window.
func TestNewActivity(t *testing.T) {
	t.Parallel()
	now := at(19, 22, 0)
	runs := []view.ActivityRun{
		{StartedAt: at(19, 7, 0), Elapsed: 2 * time.Second, State: model.StateOK},
		{StartedAt: at(19, 7, 30), Elapsed: 4 * time.Second, State: model.StateFailed},
		{StartedAt: at(19, 1, 30), Elapsed: 6 * time.Second, State: model.StateKilled}, // the 18th, 23:30 UTC
		{StartedAt: at(17, 12, 0), Elapsed: 8 * time.Second, State: model.StateStopped},
		{StartedAt: at(19, 21, 0), State: model.StateRunning},
		{StartedAt: at(19, 7, 0).AddDate(0, 0, -30), State: model.StateOK}, // a day before the window
		{StartedAt: at(20, 7, 0), State: model.StateOK},                    // tomorrow
	}

	a := view.NewActivity(runs, now, view.ActivityDays)
	assert.NotNil(t, a)
	assert.Len(t, a.Days, 30)
	assert.Len(t, a.Hours, 24)
	assert.Equal(t, 5, a.Total)
	assert.Equal(t, 2, a.Failed)
	assert.Equal(t, []string{"2026-09-20", "2026-10-05", "2026-10-19"}, a.Axis)

	today, yesterday := a.Days[29], a.Days[28]
	assert.Equal(t, "Mon 2026-10-19", today.Label)
	assert.Equal(t, 1, today.OK)
	assert.Equal(t, 2, today.Failed) // the killed run too: 01:30 here, the day before in UTC
	assert.Equal(t, 1, today.Other)
	assert.Equal(t, "4 runs, Mon 2026-10-19 · 2 failed · 1 running or stopped", today.Title)
	assert.Equal(t, 0, yesterday.OK+yesterday.Failed+yesterday.Other)
	assert.Equal(t, "no runs, Wed 2026-10-07", a.Days[17].Title)

	assert.Equal(t, 1, a.Hours[7].OK)
	assert.Equal(t, 1, a.Hours[7].Failed)
	assert.Equal(t, "2 runs, 07:00–08:00 · 1 failed", a.Hours[7].Title)
	assert.Equal(t, 1, a.Hours[1].Failed)

	// The busiest day is a full bar; its parts share it.
	assert.Equal(t, template.CSS("height:25.00%"), today.OKStyle)
	assert.Equal(t, template.CSS("height:50.00%"), today.FailedStyle)
	assert.Equal(t, template.CSS("height:25.00%"), today.OtherStyle)
	assert.Equal(t, template.CSS("height:25.00%"), a.Days[27].OtherStyle) // the 17th: stopped
	assert.Equal(t, template.CSS(""), yesterday.OKStyle)

	// Each day's and hour's bar of average duration: the mean of its
	// finished runs, the slowest full height; a running run is not in it.
	assert.Equal(t, 4*time.Second, today.Average)
	assert.Equal(t, 3, today.Finished)
	assert.Equal(t, 2*time.Second, today.Min)
	assert.Equal(t, 6*time.Second, today.Max)
	assert.Equal(t, "avg 4.0s · min 2.0s · max 6.0s over 3 runs, Mon 2026-10-19", today.AverageTitle)
	assert.Equal(t, 8*time.Second, a.Days[27].Average)
	assert.Equal(t, "8.0s, 1 run, Sat 2026-10-17", a.Days[27].AverageTitle)

	// The durations as lines against the slowest run, 8 seconds on the
	// 17th: a point in the middle of each column with finished runs, the
	// 18th without breaks them, so each is a dot.
	var lines *view.ActivityLines = a.DayLines
	assert.Equal(t, 8*time.Second, lines.Slowest)
	assert.Equal(t, "M 91.67 10.00 L 91.67 10.00 M 98.33 55.00 L 98.33 55.00", lines.Average)
	assert.Equal(t, "M 91.67 10.00 L 91.67 10.00 M 98.33 77.50 L 98.33 77.50", lines.Min)
	assert.Equal(t, "M 91.67 10.00 L 91.67 10.00 M 98.33 32.50 L 98.33 32.50", lines.Max)
	assert.Equal(t, "no finished runs, Sun 2026-10-18", yesterday.AverageTitle)
	assert.Equal(t, 3*time.Second, a.Hours[7].Average)
	assert.Equal(t, "avg 3.0s · min 2.0s · max 4.0s over 2 runs, 07:00–08:00", a.Hours[7].AverageTitle)
	assert.Equal(t, "no finished runs, 21:00–22:00", a.Hours[21].AverageTitle) // only the running one

	// The median of the finished runs: 2, 4, 6 and 8 seconds.
	assert.Equal(t, 5*time.Second, a.Median)
	assert.Len(t, a.Durations.Points, 4)

	assert.Nil(t, view.NewActivity(nil, now, view.ActivityDays))
	assert.Nil(t, view.NewActivity(runs[5:], now, view.ActivityDays))
}

// A bar with runs is never shorter than the stub, even beside a
// far busier one.
func TestActivityBar(t *testing.T) {
	t.Parallel()
	now := at(19, 22, 0)
	runs := []view.ActivityRun{{StartedAt: at(18, 9, 0), State: model.StateOK}}
	for i := 0; i < 100; i++ {
		runs = append(runs, view.ActivityRun{StartedAt: at(19, 9, 0), State: model.StateOK})
	}
	a := view.NewActivity(runs, now, 2)
	var quiet view.ActivityBar = a.Days[0]
	assert.Equal(t, template.CSS("height:4.00%"), quiet.OKStyle)
	assert.Equal(t, template.CSS("height:100.00%"), a.Days[1].OKStyle)
	assert.Len(t, a.Days, 2)
	assert.Equal(t, 101, a.Hours[9].OK)
}

// The duration line runs from the first run on the left to the last on
// the right, zero at the bottom and the longest under the headroom; each
// run hovers the stretch around it.
func TestActivitySpark(t *testing.T) {
	t.Parallel()
	now := at(19, 22, 0)
	runs := []view.ActivityRun{
		{StartedAt: at(19, 7, 0), Elapsed: 10 * time.Second, State: model.StateOK},
		{StartedAt: at(19, 8, 0), Elapsed: 5 * time.Second, State: model.StateFailed},
		{StartedAt: at(19, 9, 0), Elapsed: 0, State: model.StateOK},
	}
	var s *view.ActivitySpark = view.NewActivity(runs, now, 1).Durations
	assert.Equal(t, "M 0.00 10.00 L 50.00 55.00 L 100.00 100.00", s.Line)
	assert.Equal(t, "M 0 100 L 0.00 10.00 L 50.00 55.00 L 100.00 100.00 L 100 100 Z", s.Area)
	assert.Equal(t, 10*time.Second, s.Longest)

	var p view.ActivityPoint = s.Points[1]
	assert.Equal(t, "25.00", p.X)
	assert.Equal(t, "50.00", p.Width)
	assert.Equal(t, "50.00", p.At)
	assert.True(t, p.Failed)
	assert.Equal(t, "5.0s, Mon 2026-10-19 08:00 · failed", p.Title)
	assert.Equal(t, "0.00", s.Points[0].X)
	assert.Equal(t, "25.00", s.Points[0].Width)

	one := view.NewActivity(runs[:1], now, 1).Durations
	assert.Equal(t, "M 0 10.00 L 100 10.00", one.Line)
	assert.Equal(t, "M 0 100 L 0 10.00 L 100 10.00 L 100 100 Z", one.Area)
}

// The latest finished runs are the ones the line draws, at most
// ActivityDurations.
func TestActivity(t *testing.T) {
	t.Parallel()
	now := at(19, 22, 0)
	var runs []view.ActivityRun
	for i := 0; i < view.ActivityDurations+5; i++ {
		runs = append(runs, view.ActivityRun{StartedAt: at(19, 1, i), Elapsed: time.Second, State: model.StateOK})
	}
	var a *view.Activity = view.NewActivity(runs, now, view.ActivityDays)
	assert.Len(t, a.Durations.Points, view.ActivityDurations)
	assert.Contains(t, a.Durations.Points[0].Title, "01:05")
	assert.Equal(t, time.Second, a.Median)
}

// ActivityRun is a run as the charts read it.
func TestActivityRun(t *testing.T) {
	t.Parallel()
	run := view.ActivityRun{StartedAt: at(19, 7, 0), Elapsed: time.Minute, State: model.StateOK}
	assert.Equal(t, 1, view.NewActivity([]view.ActivityRun{run}, at(19, 8, 0), 1).Days[0].OK)
}
