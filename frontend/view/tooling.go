package view

import (
	"fmt"
	"html/template"
	"time"

	"github.com/titpetric/task-ui/tooling"
)

// ToolingWindows are the windows the panel charts, the longer first.
// Thirty days is the shape of a month, seven the shape of a week, and a
// reader comparing them sees whether this week is the usual one.
var ToolingWindows = []int{30, 7}

// toolingStub is the least height of a bar with anything in it, in
// percent, so a quiet day next to a busy one still shows.
const toolingStub = 4.0

// toolingTools is how many tools a window's bar list names. Beyond a
// handful a ranking stops ranking.
const toolingTools = 6

// Tooling is what atkins has recorded, as the runs page shows it: a
// scope per database, the project first.
type Tooling struct {
	Scopes []ToolingScope
}

// ToolingScope is one database: a chart per window, the tools behind
// each, and the commands for whoever opens the details.
type ToolingScope struct {
	// Name is what the scope is called, project or workspace, and
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

	// Commands are the normalized commands by total time, the table
	// behind the toggle.
	Commands []ToolingCommand
}

// ToolingWindow is one chart: a bar a day over its days, and the tools
// that made them.
type ToolingWindow struct {
	Days  int
	Bars  []ToolingBar
	Tools []ToolingTool

	// Runs and Total are the window's own, not the report's.
	Runs  int64
	Total time.Duration

	// Busiest is the longest day, which the chart measures against.
	Busiest time.Duration
}

// ToolingBar is one day of a chart.
type ToolingBar struct {
	Date  string
	Title string
	Style template.CSS
}

// ToolingTool is one tool of a window, with the share of the window it
// took.
type ToolingTool struct {
	Tool  string
	Runs  int64
	Total time.Duration
	// Style is the width of its bar, in percent of the busiest tool.
	Style template.CSS
}

// ToolingCommand is one normalized command of the table.
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
// atkins: the page shows no panel rather than an empty one.
func NewTooling(report *tooling.Report) *Tooling {
	scopes := report.Scopes()
	if len(scopes) == 0 {
		return nil
	}

	out := &Tooling{}
	for _, scope := range scopes {
		out.Scopes = append(out.Scopes, newToolingScope(scope))
	}
	return out
}

// newToolingScope turns one database into its section of the panel.
func newToolingScope(scope *tooling.Scope) ToolingScope {
	out := ToolingScope{
		Name:       scope.Name,
		Repository: scope.Repository,
		Database:   scope.Database,
		Days:       scope.Days,
		Runs:       scope.Runs,
		Total:      milliseconds(scope.TotalMs),
	}

	for _, days := range ToolingWindows {
		if window := newToolingWindow(scope, days); window != nil {
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
func newToolingWindow(report *tooling.Scope, days int) *ToolingWindow {
	if report.Days < days {
		return nil
	}

	window := &ToolingWindow{Days: days}

	var busiest int64
	for _, day := range report.Daily(days) {
		if day.TotalMs > busiest {
			busiest = day.TotalMs
		}
		window.Runs += day.Runs
		window.Total += milliseconds(day.TotalMs)
	}
	window.Busiest = milliseconds(busiest)

	for _, day := range report.Daily(days) {
		window.Bars = append(window.Bars, ToolingBar{
			Date:  day.Date,
			Title: toolingDayTitle(day),
			Style: heightStyle(day.TotalMs, busiest),
		})
	}

	tools := report.ToolsSince(days)
	if len(tools) > toolingTools {
		tools = tools[:toolingTools]
	}
	var widest int64
	for _, tool := range tools {
		if tool.TotalMs > widest {
			widest = tool.TotalMs
		}
	}
	for _, tool := range tools {
		window.Tools = append(window.Tools, ToolingTool{
			Tool:  tool.Tool,
			Runs:  tool.Runs,
			Total: milliseconds(tool.TotalMs),
			Style: widthStyle(tool.TotalMs, widest),
		})
	}

	return window
}

// toolingDayTitle is what a hover over a day says.
func toolingDayTitle(day tooling.Day) string {
	if day.Runs == 0 {
		return day.Date + ": nothing ran"
	}
	return fmt.Sprintf("%s: %s in %s", day.Date, plural(int(day.Runs), "command"), Duration(milliseconds(day.TotalMs)))
}

// heightStyle is the height of a bar, in percent of the busiest, with a
// floor so a quiet day beside a busy one still shows.
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

// widthStyle is the width of a bar, in percent of the widest.
func widthStyle(value, most int64) template.CSS {
	if value <= 0 || most <= 0 {
		return template.CSS("width: 0%")
	}
	width := float64(value) / float64(most) * 100
	if width < toolingStub {
		width = toolingStub
	}
	return template.CSS(fmt.Sprintf("width: %.1f%%", width))
}

// milliseconds turns the report's unit into a duration.
func milliseconds(ms int64) time.Duration {
	return time.Duration(ms) * time.Millisecond
}
