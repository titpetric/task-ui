package view_test

import (
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/tooling"
)

// toolingNow is when the page draws the panel, the evening of the report's
// last day.
var toolingNow = time.Date(2026, 10, 23, 18, 0, 0, 0, time.UTC)

// toolingReport is a project with a month of history, three days in it.
func toolingReport() *tooling.Report {
	return &tooling.Report{Project: toolingScope()}
}

// toolingScope is one database of a month, three days in it.
func toolingScope() *tooling.Scope {
	at := time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)
	return &tooling.Scope{
		Name:     "project",
		Database: "/w/.atkins/atkins.db",
		Days:     31,
		From:     "2026-09-22",
		To:       "2026-10-23",
		Runs:     44,
		TotalMs:  184_900,
		Commands: []tooling.Command{
			{Command: "go test -cover", Tool: "go", Runs: 42, TotalMs: 184_000, AverageMs: 4_380, MaxMs: 9_100, Failures: 3, LastRunAt: &at},
			{Command: "npm run build", Tool: "npm", Runs: 2, TotalMs: 900, AverageMs: 450, MaxMs: 500},
		},
		Usage: []tooling.Usage{
			{Date: "2026-10-01", Tool: "go", Runs: 10, TotalMs: 60_000},
			{Date: "2026-10-21", Tool: "go", Runs: 20, TotalMs: 90_000},
			{Date: "2026-10-23", Tool: "go", Runs: 12, TotalMs: 34_000},
			{Date: "2026-10-23", Tool: "npm", Runs: 2, TotalMs: 900},
		},
	}
}

func TestNewTooling(t *testing.T) {
	panel := view.NewTooling(toolingReport(), toolingNow)
	assert.NotNil(t, panel)
	assert.Len(t, panel.Scopes, 1)

	scope := panel.Scopes[0]
	assert.Equal(t, "project", scope.Name)
	assert.Equal(t, "/w/.atkins/atkins.db", scope.Database)
	assert.Equal(t, 31, scope.Days)
	assert.Equal(t, int64(44), scope.Runs)
	assert.Equal(t, 184_900*time.Millisecond, scope.Total)

	// A chart per window, the longer first.
	assert.Len(t, scope.Windows, 2)
	assert.Equal(t, 30, scope.Windows[0].Days)
	assert.Equal(t, 7, scope.Windows[1].Days)
	assert.Len(t, scope.Windows[0].Bars, 30)
	assert.Len(t, scope.Windows[1].Bars, 7)

	// The table behind the toggle holds every command the report
	// carries, in the order atkins ranked them.
	assert.Len(t, scope.Commands, 2)
	assert.Equal(t, "go test -cover", scope.Commands[0].Command)
	assert.Equal(t, "go", scope.Commands[0].Tool)
	assert.Equal(t, int64(3), scope.Commands[0].Failures)
	assert.Equal(t, 184_000*time.Millisecond, scope.Commands[0].Total)
	assert.NotNil(t, scope.Commands[0].LastRun)
	assert.Nil(t, scope.Commands[1].LastRun)
}

// Two databases are two sections, the project first.
func TestNewToolingBothScopes(t *testing.T) {
	workspace := toolingScope()
	workspace.Name = "workspace"
	workspace.Database = "/home/you/.atkins/atkins.db"

	panel := view.NewTooling(&tooling.Report{Project: toolingScope(), Workspace: workspace}, toolingNow)
	assert.NotNil(t, panel)
	assert.Len(t, panel.Scopes, 2)
	assert.Equal(t, "project", panel.Scopes[0].Name)
	assert.Equal(t, "workspace", panel.Scopes[1].Name)
	assert.Equal(t, "/home/you/.atkins/atkins.db", panel.Scopes[1].Database)
}

// A tree that has never run a job through atkins shows no panel rather
// than an empty one.
func TestNewToolingWithNothingRecorded(t *testing.T) {
	assert.Nil(t, view.NewTooling(nil, toolingNow))
	assert.Nil(t, view.NewTooling(&tooling.Report{}, toolingNow))
	assert.Nil(t, view.NewTooling(&tooling.Report{Workspace: &tooling.Scope{Database: "/w/x.db", Days: 31}}, toolingNow))
}

// A report shorter than a window draws no chart for it: a week of
// history drawn as a month is a month of empty bars with a week on the
// end.
func TestNewToolingSkipsAWindowItCannotFill(t *testing.T) {
	report := toolingReport()
	report.Project.Days = 7

	panel := view.NewTooling(report, toolingNow)
	assert.NotNil(t, panel)
	assert.Len(t, panel.Scopes[0].Windows, 1)
	assert.Equal(t, 7, panel.Scopes[0].Windows[0].Days)
}

// Every day is a group of bars, a slot per legend entry, each measured
// from zero against the window's tallest bar; a command that did not
// run that day leaves its slot empty, and a day nothing ran has none.
func TestToolingBars(t *testing.T) {
	window := view.NewTooling(toolingReport(), toolingNow).Scopes[0].Windows[1]

	assert.Equal(t, 7, len(window.Bars))
	assert.Equal(t, "2026-10-17", window.Bars[0].Date)
	assert.Equal(t, "2026-10-23", window.Bars[6].Date)
	assert.Equal(t, 90_000*time.Millisecond, window.Tallest)

	// The 21st ran go alone, the tallest bar of the week, and npm's
	// slot beside it stays empty.
	day := window.Bars[4]
	assert.Len(t, day.Segments, 2)
	assert.Equal(t, "height: 100.0%", string(day.Segments[0].Style))
	assert.Equal(t, "npm", day.Segments[1].Name)
	assert.Equal(t, "", string(day.Segments[1].Style))
	assert.Equal(t, "", day.Segments[1].Title)

	// The 23rd ran both, each its share of the tallest, the quiet one
	// on the floor.
	assert.Contains(t, string(window.Bars[6].Segments[0].Style), "height: 37.")
	assert.Equal(t, "height: 4.0%", string(window.Bars[6].Segments[1].Style))

	// A day nothing ran draws nothing, and says so on hover.
	assert.Len(t, window.Bars[0].Segments, 0)
	assert.Equal(t, "2026-10-17: nothing ran", window.Bars[0].Title)

	// A day that ran says what and for how long.
	assert.Contains(t, window.Bars[4].Title, "2026-10-21: 20 commands in ")
	assert.Equal(t, 90_000*time.Millisecond, window.Busiest)
}

// The chart draws a bar with anything in it at least at the floor, so
// a quiet command beside a busy one still shows.
func TestToolingBarsFloor(t *testing.T) {
	report := toolingReport()
	report.Project.Usage = append(report.Project.Usage, tooling.Usage{Date: "2026-10-22", Tool: "go", Runs: 1, TotalMs: 1})

	window := view.NewTooling(report, toolingNow).Scopes[0].Windows[1]
	assert.Equal(t, "height: 4.0%", string(window.Bars[5].Segments[0].Style))
	assert.Contains(t, window.Bars[5].Title, "1 command in ")
}

// Without a command series, from an older atkins, the bars are of the
// tools and the legend names them, rolled up from the window's days.
func TestToolingStacksByToolWithoutCommandUsage(t *testing.T) {
	window := view.NewTooling(toolingReport(), toolingNow).Scopes[0].Windows[1]
	assert.Equal(t, "tool", window.By)

	assert.Len(t, window.Legend, 2)
	assert.Equal(t, "go", window.Legend[0].Name)
	assert.Equal(t, int64(32), window.Legend[0].Runs)
	assert.Equal(t, 124_000*time.Millisecond, window.Legend[0].Total)
	assert.Equal(t, "s0", window.Legend[0].Color)
	assert.Equal(t, "npm", window.Legend[1].Name)
	assert.Equal(t, "s1", window.Legend[1].Color)

	// The 23rd ran both, go first.
	day := window.Bars[6]
	assert.Len(t, day.Segments, 2)
	assert.Equal(t, "go", day.Segments[0].Name)
	assert.Equal(t, "npm", day.Segments[1].Name)
	assert.Equal(t, "tip-end", day.Edge)

	// The totals count the window's days alone.
	assert.Equal(t, int64(34), window.Runs)
	assert.Equal(t, 124_900*time.Millisecond, window.Total)
}

// commandScope is toolingScope with a command series of seven keys on
// the last day, so the top five stack apart and two make up other.
func commandScope() *tooling.Scope {
	scope := toolingScope()
	for index, command := range []string{"go test", "docker build", "docker compose up", "git clone", "npm run build", "go vet", "make"} {
		scope.CommandUsage = append(scope.CommandUsage, tooling.CommandUsage{
			Date: "2026-10-23", Command: command, Tool: strings.Fields(command)[0],
			Runs: 1, TotalMs: int64(7000 - 1000*index),
		})
	}
	return scope
}

// With a command series each day draws a bar per top five command of
// the window, the rest together as other, and the legend names them.
func TestToolingStacksTopCommands(t *testing.T) {
	window := view.NewTooling(&tooling.Report{Project: commandScope()}, toolingNow).Scopes[0].Windows[1]
	assert.Equal(t, "command", window.By)

	assert.Len(t, window.Legend, 6)
	assert.Equal(t, "go test", window.Legend[0].Name)
	assert.Equal(t, "docker compose up", window.Legend[2].Name)
	assert.Equal(t, "npm run build", window.Legend[4].Name)
	assert.Equal(t, "other", window.Legend[5].Name)
	assert.Equal(t, "s-other", window.Legend[5].Color)
	assert.Equal(t, int64(2), window.Legend[5].Runs)
	assert.Equal(t, 3000*time.Millisecond, window.Legend[5].Total)

	day := window.Bars[6]
	assert.Len(t, day.Segments, 6)
	assert.Equal(t, "go test", day.Segments[0].Name)
	assert.Equal(t, "s0", day.Segments[0].Color)
	assert.Equal(t, "height: 100.0%", string(day.Segments[0].Style))
	assert.Equal(t, "2026-10-23 · go test: 7.0s, 1 run", day.Segments[0].Title)
	assert.Equal(t, "height: 85.7%", string(day.Segments[1].Style))
	assert.Equal(t, "other", day.Segments[5].Name)
	assert.Equal(t, "s-other", day.Segments[5].Color)
	assert.Equal(t, "2026-10-23 · other: 3.0s, 2 runs", day.Segments[5].Title)

	// A day nothing ran has no segments.
	assert.Len(t, window.Bars[0].Segments, 0)
	assert.Equal(t, "tip-start", window.Bars[0].Edge)
}

// A command keeps its colour in every window and scope of the panel.
func TestToolingColorsAreStable(t *testing.T) {
	workspace := commandScope()
	workspace.Name = "workspace"
	workspace.CommandUsage = append(workspace.CommandUsage, tooling.CommandUsage{Date: "2026-10-01", Command: "make", Tool: "make", Runs: 1, TotalMs: 99_000})

	panel := view.NewTooling(&tooling.Report{Project: commandScope(), Workspace: workspace}, toolingNow)
	colors := map[string]string{}
	for _, scope := range panel.Scopes {
		for _, window := range scope.Windows {
			for _, entry := range window.Legend {
				if seen, ok := colors[entry.Name]; ok {
					assert.Equal(t, seen, entry.Color, entry.Name)
				}
				colors[entry.Name] = entry.Color
			}
		}
	}
	// The workspace's month makes make a top command, in a colour of
	// its own.
	assert.Equal(t, "s5", colors["make"])
}

// The panel renders inside the runs page without a template error, and
// what it draws is the report's own numbers.
func TestToolingRenders(t *testing.T) {
	v, err := view.New()
	assert.NoError(t, err)

	var out strings.Builder
	page := &view.Page{Title: "Runs", Tooling: view.NewTooling(toolingReport(), toolingNow)}
	assert.NoError(t, v.Render(&out, "runs", page))

	html := out.String()
	assert.Contains(t, html, "Tooling")
	assert.Contains(t, html, "Last 30 days")
	assert.Contains(t, html, "Last 7 days")
	assert.Contains(t, html, "2026-10-21: 20 commands in ")
	assert.Contains(t, html, "go test -cover")
	assert.Contains(t, html, "/w/.atkins/atkins.db")
	assert.Contains(t, html, "project")
	// The details element is the toggle, and it needs no script.
	assert.Contains(t, html, "<details")
	assert.Contains(t, html, "by the time they have taken")
	assert.Contains(t, html, "tooling-legend")
	assert.Contains(t, html, `<span class="swatch s0"></span><span class="key" title="go">go</span>`)
	assert.Contains(t, html, `<span style="left: 92.86%">Fri</span>`)
	// The 21st ran go alone: npm's slot beside it stays bare.
	assert.Contains(t, html, `data-tip="2026-10-21 · go: 1m30s, 20 runs"></b><b></b></span>`)
	// The details toggle wears the shared disclosure chevron.
	assert.Contains(t, html, `<summary><svg class="ic chev"><use href="#i-chev"/></svg>2 commands by the time they have taken</summary>`)
}

// The axis marks every day of the week and a day a week of the month,
// the last day last.
func TestToolingTicks(t *testing.T) {
	t.Parallel()

	scope := view.NewTooling(toolingReport(), toolingNow).Scopes[0]

	week := scope.Windows[1].Ticks
	assert.Len(t, week, 7)
	assert.Equal(t, view.ActivityTick{Label: "Fri", At: "92.86"}, week[6])
	assert.Equal(t, view.ActivityTick{Label: "Sat", At: "7.14"}, week[0])

	month := scope.Windows[0].Ticks
	assert.Len(t, month, 5)
	assert.Equal(t, view.ActivityTick{Label: "Oct 23", At: "98.33"}, month[4])
	assert.Equal(t, view.ActivityTick{Label: "Sep 25", At: "5.00"}, month[0])
}

// The Activity and Tooling blocks of the runs page each fold on a
// toggle in their heading, open by default.
func TestToolingRendersFoldToggles(t *testing.T) {
	v, err := view.New()
	assert.NoError(t, err)

	var out strings.Builder
	page := &view.Page{Title: "Runs", Tooling: view.NewTooling(&tooling.Report{Project: commandScope()}, toolingNow)}
	assert.NoError(t, v.Render(&out, "runs", page))

	html := out.String()
	assert.Contains(t, html, `data-fold="runs-tooling"`)
	// The fold toggle wears the shared disclosure chevron, the one the
	// sidebar's workspaces, projects and folders fold on.
	assert.Contains(t, html, `class="fold-toggle" aria-expanded="true" aria-controls="runs-tooling"><svg class="ic chev"><use href="#i-chev"/></svg>Tooling</button>`)
	assert.Contains(t, html, `id="runs-tooling"`)
	assert.Contains(t, html, "docker compose up")
	assert.Contains(t, html, `data-tip="2026-10-23 · go test: 7.0s, 1 run"`)
	// A day's group holds a slot per legend entry.
	assert.Contains(t, html, `style="--n: 7; --k: 6"`)
}

// A page with nothing recorded draws no panel at all.
func TestToolingRendersNothing(t *testing.T) {
	v, err := view.New()
	assert.NoError(t, err)

	var out strings.Builder
	assert.NoError(t, v.Render(&out, "runs", &view.Page{Title: "Runs"}))
	assert.NotContains(t, out.String(), "tooling-charts")
}

// TestTooling covers the panel as a whole: what it says about the tree
// above its charts.
func TestTooling(t *testing.T) {
	var panel view.Tooling
	assert.Len(t, panel.Scopes, 0)

	panel = *view.NewTooling(toolingReport(), toolingNow)
	assert.Len(t, panel.Scopes, 1)
}

// TestToolingWindow covers one chart: its days, its bars and its tools.
func TestToolingWindow(t *testing.T) {
	var window view.ToolingWindow
	assert.Equal(t, 0, window.Days)

	window = view.NewTooling(toolingReport(), toolingNow).Scopes[0].Windows[0]
	assert.Equal(t, 30, window.Days)
	assert.Len(t, window.Bars, 30)
	assert.NotNil(t, window.Legend)
}

// TestToolingBar covers one day of a chart.
func TestToolingBar(t *testing.T) {
	bar := view.NewTooling(toolingReport(), toolingNow).Scopes[0].Windows[1].Bars[4]
	assert.Equal(t, "2026-10-21", bar.Date)
	assert.Contains(t, bar.Title, "20 commands")
	assert.Len(t, bar.Segments, 2)

	// A day nobody drew has no slots, and the template draws an empty
	// column.
	assert.Len(t, view.ToolingBar{Date: "2026-10-22"}.Segments, 0)
}

// TestToolingCommand covers one row of the table behind the toggle.
func TestToolingCommand(t *testing.T) {
	command := view.NewTooling(toolingReport(), toolingNow).Scopes[0].Commands[0]
	assert.Equal(t, "go test -cover", command.Command)
	assert.Equal(t, "go", command.Tool)
	assert.Equal(t, int64(3), command.Failures)
	assert.Equal(t, 9_100*time.Millisecond, command.Max)

	assert.Nil(t, view.ToolingCommand{}.LastRun)
}

// TestToolingScope covers one database's section of the panel.
func TestToolingScope(t *testing.T) {
	var scope view.ToolingScope
	assert.Equal(t, "", scope.Name)

	workspace := toolingScope()
	workspace.Name = "workspace"
	workspace.Repository = ""

	panel := view.NewTooling(&tooling.Report{Project: toolingScope(), Workspace: workspace}, toolingNow)
	assert.Equal(t, "project", panel.Scopes[0].Name)
	assert.Equal(t, 31, panel.Scopes[0].Days)
	assert.Len(t, panel.Scopes[0].Windows, 2)
	assert.Equal(t, "", panel.Scopes[1].Repository)
}

// Both sections render, each with its own heading, charts and table.
func TestToolingRendersBothScopes(t *testing.T) {
	v, err := view.New()
	assert.NoError(t, err)

	project := toolingScope()
	project.Repository = "titpetric/atkins"
	workspace := toolingScope()
	workspace.Name = "workspace"
	workspace.Database = "/home/you/.atkins/atkins.db"

	var out strings.Builder
	page := &view.Page{Title: "Runs", Tooling: view.NewTooling(&tooling.Report{Project: project, Workspace: workspace}, toolingNow)}
	assert.NoError(t, v.Render(&out, "runs", page))

	html := out.String()
	assert.Equal(t, 2, strings.Count(html, "tooling-scope"))
	assert.Contains(t, html, "titpetric/atkins")
	assert.Contains(t, html, "/home/you/.atkins/atkins.db")
	assert.Equal(t, 2, strings.Count(html, "<details"))
}

// More commands than colours across the windows: each window's legend
// still holds a colour once, and a command keeps its colour.
func TestToolingColorsAreDistinctInAWindow(t *testing.T) {
	t.Parallel()

	scope := toolingScope()
	for index, command := range []string{"a", "b", "c", "d", "e"} {
		scope.CommandUsage = append(scope.CommandUsage, tooling.CommandUsage{Date: "2026-10-01", Command: command, Tool: command, Runs: 1, TotalMs: int64(90_000 - 1000*index)})
	}
	for _, command := range []string{"f", "g", "h"} {
		scope.CommandUsage = append(scope.CommandUsage, tooling.CommandUsage{Date: "2026-10-23", Command: command, Tool: command, Runs: 1, TotalMs: 500})
	}

	windows := view.NewTooling(&tooling.Report{Project: scope}, toolingNow).Scopes[0].Windows
	for _, window := range windows {
		seen := map[string]string{}
		for _, entry := range window.Legend {
			other, ok := seen[entry.Color]
			assert.False(t, ok, entry.Name+" shares "+entry.Color+" with "+other)
			seen[entry.Color] = entry.Name
		}
	}
	assert.Equal(t, "s5", windows[1].Legend[0].Color)
	assert.Equal(t, "f", windows[1].Legend[0].Name)
}
