package tooling_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/tooling"
)

// report is three days of two tools, the last of them the day the
// report ends on.
func report() *tooling.Report {
	return &tooling.Report{
		Database: "/w/.atkins/atkins.db",
		Days:     31,
		From:     "2026-09-22",
		To:       "2026-10-23",
		Runs:     44,
		TotalMs:  184_900,
		Commands: []tooling.Command{
			{Command: "go test -cover", Tool: "go", Runs: 42, TotalMs: 184_000},
		},
		Tools: []tooling.Tool{
			{Tool: "go", Commands: 1, Runs: 42, TotalMs: 184_000},
			{Tool: "npm", Commands: 1, Runs: 2, TotalMs: 900},
		},
		Usage: []tooling.Usage{
			{Date: "2026-10-17", Tool: "go", Runs: 10, TotalMs: 60_000},
			{Date: "2026-10-21", Tool: "go", Runs: 20, TotalMs: 90_000},
			{Date: "2026-10-23", Tool: "go", Runs: 12, TotalMs: 34_000},
			{Date: "2026-10-23", Tool: "npm", Runs: 2, TotalMs: 900},
		},
	}
}

func TestReport_Empty(t *testing.T) {
	var none *tooling.Report
	assert.True(t, none.Empty())
	assert.True(t, (&tooling.Report{}).Empty())

	// A report with a database and a window but nothing recorded is
	// still nothing to draw.
	assert.True(t, (&tooling.Report{Database: "/w/x.db", Days: 31}).Empty())
	assert.False(t, report().Empty())
}

func TestReport_Since(t *testing.T) {
	// The window counts back from the day the report ends on, and
	// includes it: seven days ending on the 23rd starts on the 17th.
	assert.Len(t, report().Since(1), 2)
	assert.Len(t, report().Since(3), 3)
	assert.Len(t, report().Since(5), 3)
	assert.Len(t, report().Since(7), 4)
	assert.Len(t, report().Since(30), 4)

	assert.Nil(t, report().Since(0))
	assert.Nil(t, report().Since(-1))

	var none *tooling.Report
	assert.Nil(t, none.Since(7))

	// A report whose last day is not a date has no window to count
	// back from.
	assert.Nil(t, (&tooling.Report{To: "never", Usage: report().Usage}).Since(7))
}

func TestReport_ToolsSince(t *testing.T) {
	// A week rolls up from the usage rows rather than from the
	// report's own tools, which cover its whole window.
	week := report().ToolsSince(7)
	assert.Len(t, week, 2)
	assert.Equal(t, "go", week[0].Tool)
	assert.Equal(t, int64(184_000), week[0].TotalMs)
	assert.Equal(t, int64(42), week[0].Runs)
	assert.Equal(t, "npm", week[1].Tool)
	assert.Equal(t, int64(900), week[1].TotalMs)

	// A shorter window leaves out the days before it.
	shorter := report().ToolsSince(5)
	assert.Equal(t, int64(124_000), shorter[0].TotalMs)
	assert.Equal(t, int64(32), shorter[0].Runs)

	// A day holds one tool of the two.
	day := report().ToolsSince(1)
	assert.Len(t, day, 2)
	assert.Equal(t, "go", day[0].Tool)
	assert.Equal(t, int64(34_000), day[0].TotalMs)

	assert.Len(t, report().ToolsSince(0), 0)
}

// Two tools that cost the same are ordered by name, so the list does
// not shuffle between reloads.
func TestReport_ToolsSinceTies(t *testing.T) {
	tied := &tooling.Report{
		Days: 31, To: "2026-10-23",
		Usage: []tooling.Usage{
			{Date: "2026-10-23", Tool: "npm", Runs: 1, TotalMs: 100},
			{Date: "2026-10-23", Tool: "docker", Runs: 1, TotalMs: 100},
			{Date: "2026-10-23", Tool: "go", Runs: 1, TotalMs: 100},
		},
	}

	var names []string
	for _, tool := range tied.ToolsSince(7) {
		names = append(names, tool.Tool)
	}
	assert.Equal(t, []string{"docker", "go", "npm"}, names)
}

func TestReport_Daily(t *testing.T) {
	days := report().Daily(7)
	assert.Len(t, days, 7)

	// Oldest first, ending on the report's last day.
	assert.Equal(t, "2026-10-17", days[0].Date)
	assert.Equal(t, "2026-10-23", days[6].Date)

	// A day nothing ran is a zero and not a gap.
	assert.Equal(t, int64(60_000), days[0].TotalMs)
	assert.Equal(t, int64(0), days[1].TotalMs)
	assert.Equal(t, int64(0), days[1].Runs)

	// Both tools of a day fold into its one total.
	assert.Equal(t, int64(34_900), days[6].TotalMs)
	assert.Equal(t, int64(14), days[6].Runs)

	assert.Nil(t, report().Daily(0))

	var none *tooling.Report
	assert.Nil(t, none.Daily(7))
	assert.Nil(t, (&tooling.Report{To: "never"}).Daily(7))
}

// A report of thirty days draws thirty bars, so the chart is the width
// of its window whatever is in it.
func TestReport_DailyFillsTheWindow(t *testing.T) {
	days := report().Daily(30)
	assert.Len(t, days, 30)
	assert.Equal(t, "2026-09-24", days[0].Date)
	assert.Equal(t, "2026-10-23", days[29].Date)

	var ran int
	for _, day := range days {
		if day.Runs > 0 {
			ran++
		}
	}
	assert.Equal(t, 3, ran)
}

// The report's window is the one atkins keeps, which the panel reads
// to decide which charts it can draw.
func TestReport(t *testing.T) {
	r := report()
	assert.Equal(t, 31, r.Days)
	assert.Equal(t, "/w/.atkins/atkins.db", r.Database)

	last, err := time.Parse(time.DateOnly, r.To)
	assert.NoError(t, err)
	assert.Equal(t, 2026, last.Year())
}

// TestCommand covers one normalized command as the report carries it.
func TestCommand(t *testing.T) {
	command := report().Commands[0]
	assert.Equal(t, "go test -cover", command.Command)
	assert.Equal(t, "go", command.Tool)
	assert.Equal(t, int64(42), command.Runs)
	assert.Equal(t, int64(184_000), command.TotalMs)
}

// TestTool covers one tool rolled up across every form of it.
func TestTool(t *testing.T) {
	tool := report().Tools[0]
	assert.Equal(t, "go", tool.Tool)
	assert.Equal(t, int64(1), tool.Commands)
	assert.Equal(t, int64(42), tool.Runs)
}

// TestUsage covers one tool on one day, the row a chart is drawn from.
func TestUsage(t *testing.T) {
	usage := report().Usage[0]
	assert.Equal(t, "2026-10-17", usage.Date)
	assert.Equal(t, "go", usage.Tool)
	assert.Equal(t, int64(10), usage.Runs)
	assert.Equal(t, int64(60_000), usage.TotalMs)
}

// TestDay covers one day of a window, as Daily fills it.
func TestDay(t *testing.T) {
	day := report().Daily(1)[0]
	assert.Equal(t, "2026-10-23", day.Date)
	assert.Equal(t, int64(14), day.Runs)
	assert.Equal(t, int64(34_900), day.TotalMs)

	empty := tooling.Day{Date: "2026-10-22"}
	assert.Equal(t, int64(0), empty.Runs)
}
