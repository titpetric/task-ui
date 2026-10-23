package view

import (
	"fmt"
	"html/template"
	"sort"
	"time"

	"github.com/titpetric/task-ui/tooling"
)

// ToolingWindows are the windows the panel charts, the longer first.
// Thirty days is the shape of a month, seven the shape of a week, and a
// reader comparing them sees whether this week is the usual one.
var ToolingWindows = []int{30, 7}

// toolingStub is the least height of a bar with anything in it, in
// percent, so a quiet command next to a busy one still shows.
const toolingStub = 4.0

// toolingTop is how many commands a window draws a bar each for, the
// rest drawn together as "other". Beyond a handful of colours a day's
// group stops reading as one.
const toolingTop = 5

// toolingColors is how many series colours the palette holds.
const toolingColors = 6

// toolingOther is the colour class of the commands past the top ones.
const toolingOther = "s-other"

// Tooling is what atkins has recorded, as the runs page shows it: a
// scope per database, the project first.
type Tooling struct {
	Scopes []ToolingScope
}

// ToolingScope is one database: a chart per window, the tools behind
// each, and the commands for whoever opens the details.
type ToolingScope struct {
	// Name names the scope, project or workspace, and
	// Repository the checkout a project's rows came from.
	Name       string
	Repository string

	// Database is the file the rows came from, and Days the window
	// atkins keeps, which bounds every chart here.
	Database string
	Days     int

	// Runs is how many commands ran in that window and Total how long
	// they took between them.
	Runs  int64
	Total time.Duration

	// Windows are the charts, the longer first.
	Windows []ToolingWindow

	// Commands are the command keys by total time, the table behind
	// the toggle.
	Commands []ToolingCommand
}

// ToolingWindow is one chart: a group of bars a day over its days, a
// bar per top command of the window, and the legend that names them.
type ToolingWindow struct {
	Days   int
	Bars   []ToolingBar
	Legend []ToolingLegend

	// Ticks mark the days under the chart: a day a week on the month,
	// every day on the week.
	Ticks []ActivityTick

	// By is what the bars are of: "command", or "tool" from an atkins
	// that reports no command series.
	By string

	// Runs and Total count the window's days alone.
	Runs  int64
	Total time.Duration

	// Busiest is the longest day, and Tallest the longest one command
	// took in a day, which every bar measures against.
	Busiest time.Duration
	Tallest time.Duration
}

// ToolingBar is one day of a chart: a slot per entry of the legend, in
// its order, each a bar against the window's tallest, empty for a
// command that did not run that day. A day nothing ran has no slots.
type ToolingBar struct {
	Date     string
	Title    string
	Segments []ToolingSegment

	// Edge anchors the hover near either end of the chart, so it stays
	// inside it: tip-start, tip-end or nothing.
	Edge string
}

// ToolingCommand is one command key of the table.
type ToolingCommand struct {
	Command  string
	Tool     string
	Runs     int64
	Total    time.Duration
	Average  time.Duration
	Max      time.Duration
	Failures int64
	LastRun  *time.Time
}

// NewTooling turns a report into the panel. Nil for a report with
// nothing in it, which is a tree that has never run a job through
// atkins: the page then shows no panel at all.
func NewTooling(report *tooling.Report) *Tooling {
	scopes := report.Scopes()
	if len(scopes) == 0 {
		return nil
	}

	out := &Tooling{}
	palette := toolingPalette{}
	for _, scope := range scopes {
		out.Scopes = append(out.Scopes, newToolingScope(scope, palette))
	}
	return out
}

// toolingPalette keeps a command's colour across the windows and the
// scopes of the panel: each new command takes the colour fewest names
// hold that no other name of its window does, so two commands of one
// legend never share one.
type toolingPalette map[string]string

// color returns the colour class of a name, and sets a name it has
// not seen to a colour none of taken holds, taken the colours of the
// names beside it in its window.
func (p toolingPalette) color(name string, taken map[string]bool) string {
	if color, ok := p[name]; ok {
		return color
	}
	held := map[string]int{}
	for _, color := range p {
		held[color]++
	}
	best := ""
	for index := range toolingColors {
		color := fmt.Sprintf("s%d", index)
		if taken[color] {
			continue
		}
		if best == "" || held[color] < held[best] {
			best = color
		}
	}
	if best == "" {
		best = fmt.Sprintf("s%d", len(p)%toolingColors)
	}
	p[name] = best
	return best
}

// newToolingScope turns one database into its section of the panel.
func newToolingScope(scope *tooling.Scope, palette toolingPalette) ToolingScope {
	out := ToolingScope{
		Name:       scope.Name,
		Repository: scope.Repository,
		Database:   scope.Database,
		Days:       scope.Days,
		Runs:       scope.Runs,
		Total:      milliseconds(scope.TotalMs),
	}

	for _, days := range ToolingWindows {
		if window := newToolingWindow(scope, days, palette); window != nil {
			out.Windows = append(out.Windows, *window)
		}
	}

	for _, command := range scope.Commands {
		out.Commands = append(out.Commands, ToolingCommand{
			Command:  command.Command,
			Tool:     command.Tool,
			Runs:     command.Runs,
			Total:    milliseconds(command.TotalMs),
			Average:  milliseconds(command.AverageMs),
			Max:      milliseconds(command.MaxMs),
			Failures: command.Failures,
			LastRun:  command.LastRunAt,
		})
	}

	return out
}

// newToolingWindow builds one chart. Nil when the report is shorter
// than the window: a week of history drawn as a month is a month of
// empty bars with a week at the end.
func newToolingWindow(report *tooling.Scope, days int, palette toolingPalette) *ToolingWindow {
	if report.Days < days {
		return nil
	}

	daily := report.Daily(days)
	window := &ToolingWindow{Days: days}

	var busiest int64
	for _, day := range daily {
		busiest = max(busiest, day.TotalMs)
		window.Runs += day.Runs
		window.Total += milliseconds(day.TotalMs)
	}
	window.Busiest = milliseconds(busiest)

	rows, by := toolingRows(report, daily)
	window.By = by

	// The top names by the window's total, which the legend lists and
	// each day draws a bar of, in that order.
	totals := map[string]*ToolingLegend{}
	var names []string
	for _, row := range rows {
		entry, ok := totals[row.name]
		if !ok {
			entry = &ToolingLegend{Name: row.name}
			totals[row.name] = entry
			names = append(names, row.name)
		}
		entry.Runs += row.runs
		entry.Total += milliseconds(row.ms)
	}
	sort.SliceStable(names, func(i, j int) bool {
		if totals[names[i]].Total != totals[names[j]].Total {
			return totals[names[i]].Total > totals[names[j]].Total
		}
		return names[i] < names[j]
	})

	top := names[:min(len(names), toolingTop)]
	rank := map[string]int{}
	// The names that have a colour keep it; the rest take one none of
	// those holds.
	taken := map[string]bool{}
	for _, name := range top {
		if color, ok := palette[name]; ok {
			taken[color] = true
		}
	}
	for index, name := range top {
		rank[name] = index
		entry := totals[name]
		entry.Color = palette.color(name, taken)
		taken[entry.Color] = true
		window.Legend = append(window.Legend, *entry)
	}
	if len(names) > len(top) {
		other := ToolingLegend{Name: "other", Color: toolingOther}
		for _, name := range names[len(top):] {
			other.Runs += totals[name].Runs
			other.Total += totals[name].Total
		}
		window.Legend = append(window.Legend, other)
	}

	// A day's rows summed per slot: the top names, then other.
	slots := map[string][]toolingRow{}
	var tallest int64
	for _, row := range rows {
		day, ok := slots[row.date]
		if !ok {
			day = make([]toolingRow, len(window.Legend))
			slots[row.date] = day
		}
		index, ok := rank[row.name]
		if !ok {
			index = len(top)
		}
		day[index].runs += row.runs
		day[index].ms += row.ms
		tallest = max(tallest, day[index].ms)
	}
	window.Tallest = milliseconds(tallest)

	for index, day := range daily {
		bar := ToolingBar{
			Date:  day.Date,
			Title: toolingDayTitle(day),
		}
		switch {
		case index < 4:
			bar.Edge = "tip-start"
		case index >= len(daily)-4:
			bar.Edge = "tip-end"
		}
		bar.Segments = toolingSegments(day, slots[day.Date], window.Legend, tallest)
		window.Bars = append(window.Bars, bar)
	}
	window.Ticks = toolingTicks(daily)

	return window
}

// toolingTicks marks the days under a chart, centred on their column:
// on a week every day by its weekday, on a longer window a day a week,
// the last day the last, by its date.
func toolingTicks(daily []tooling.Day) []ActivityTick {
	var ticks []ActivityTick
	n := len(daily)
	step, layout := 7, "Jan 2"
	if n <= 7 {
		step, layout = 1, "Mon"
	}
	for i := n - 1; i >= 0; i -= step {
		day, err := time.Parse(time.DateOnly, daily[i].Date)
		if err != nil {
			continue
		}
		at := svgNumber((float64(i) + 0.5) * 100 / float64(n))
		ticks = append([]ActivityTick{{Label: day.Format(layout), At: at}}, ticks...)
	}
	return ticks
}

// toolingRow is one name on one day, a command or a tool.
type toolingRow struct {
	date, name string
	runs, ms   int64
}

// toolingRows returns the rows of the window's days, by command when
// atkins reports the command series and by tool when it does not.
func toolingRows(report *tooling.Scope, daily []tooling.Day) ([]toolingRow, string) {
	if len(daily) == 0 {
		return nil, "command"
	}
	first := daily[0].Date

	var rows []toolingRow
	if len(report.CommandUsage) > 0 {
		for _, row := range report.CommandUsage {
			if row.Date >= first {
				rows = append(rows, toolingRow{date: row.Date, name: row.Command, runs: row.Runs, ms: row.TotalMs})
			}
		}
		return rows, "command"
	}

	for _, row := range report.Usage {
		if row.Date >= first {
			rows = append(rows, toolingRow{date: row.Date, name: row.Tool, runs: row.Runs, ms: row.TotalMs})
		}
	}
	return rows, "tool"
}

// toolingSegments is one day's group: a slot per legend entry, in its
// order, each a bar against the window's tallest. A slot of a command
// that did not run that day is empty and says nothing; a day nothing
// ran has no slots, and its column says so.
func toolingSegments(day tooling.Day, slots []toolingRow, legend []ToolingLegend, tallest int64) []ToolingSegment {
	var sum int64
	for _, slot := range slots {
		sum += slot.ms
	}
	if sum <= 0 {
		return nil
	}

	segments := make([]ToolingSegment, len(legend))
	for index, entry := range legend {
		segments[index] = ToolingSegment{Name: entry.Name, Color: entry.Color}
		slot := slots[index]
		if slot.ms <= 0 {
			continue
		}
		segments[index].Title = fmt.Sprintf("%s · %s: %s, %s", day.Date, entry.Name, Duration(milliseconds(slot.ms)), plural(int(slot.runs), "run"))
		segments[index].Style = heightStyle(slot.ms, tallest)
	}
	return segments
}

// toolingDayTitle is what a hover over a day says.
func toolingDayTitle(day tooling.Day) string {
	if day.Runs == 0 {
		return day.Date + ": nothing ran"
	}
	return fmt.Sprintf("%s: %s in %s", day.Date, plural(int(day.Runs), "command"), Duration(milliseconds(day.TotalMs)))
}

// heightStyle is the height of a bar, in percent of the tallest, with a
// floor so a quiet command beside a busy one still shows.
func heightStyle(value, most int64) template.CSS {
	if value <= 0 || most <= 0 {
		return ""
	}
	height := float64(value) / float64(most) * 100
	if height < toolingStub {
		height = toolingStub
	}
	return template.CSS(fmt.Sprintf("height: %.1f%%", height))
}

// milliseconds turns the report's unit into a duration.
func milliseconds(ms int64) time.Duration {
	return time.Duration(ms) * time.Millisecond
}
