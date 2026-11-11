// Package tooling asks atkins what a checkout has been running.
//
// `atkins --stats --json --all` reports the commands a run executed,
// rolled up per tool and per day and listed one by one, over the window
// atkins keeps. task-ui reads it
// to answer a question its own run history cannot: a run of a job is one
// row here, and the twenty commands inside it are twenty rows there, so
// the time a repository spends on go, npm or docker is atkins' to
// report and not task-ui's to guess.
//
// Nothing in this package runs a command of its own. It execs atkins,
// decodes what it prints, and keeps the result for a while.
package tooling

import (
	"sort"
	"strings"
	"time"
)

// Report is what atkins says about a directory: what this checkout has
// been running, and what every checkout on the machine has been running
// between them.
//
// A run records to both databases, so the two answer different
// questions of the same rows. The project's says what this checkout
// costs; the workspace's says where a week went across every project
// atkins ran in.
type Report struct {
	// Project is the checkout's own history, absent for a directory
	// that belongs to no project.
	Project *Scope `json:"project,omitempty"`

	// Workspace is every project's history together.
	Workspace *Scope `json:"workspace,omitempty"`
}

// Scopes returns the scopes the report holds, the project first.
func (r *Report) Scopes() []*Scope {
	if r == nil {
		return nil
	}
	var scopes []*Scope
	for _, scope := range []*Scope{r.Project, r.Workspace} {
		if !scope.Empty() {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}

// Empty reports a report with nothing recorded in either database.
func (r *Report) Empty() bool {
	return len(r.Scopes()) == 0
}

// Scope is one database's history.
//
// Since and the methods built on it count a window in days back from the
// report's last day: the rows are dates, and a reader asks "the last
// seven days" of a calendar. The
// report's own Tools cover its whole window, so ToolsSince rolls a
// shorter one up from the usage rows. Daily fills the empty days with
// zeros: a chart that leaves them out draws a busy fortnight and a quiet
// one the same width.
type Scope struct {
	// Name is the scope's name, project or workspace.
	Name string `json:"name"`

	// Repository is the checkout the rows came from, empty for the
	// workspace, which holds many.
	Repository string `json:"repository,omitempty"`

	// Database is the file the rows came from, and Days the window they
	// cover.
	Database string `json:"database"`
	Days     int    `json:"days"`

	// From and To bound the window, as dates.
	From string `json:"from"`
	To   string `json:"to"`

	// Runs is how many commands ran in the window and TotalMs how long
	// they took between them.
	Runs    int64 `json:"runs"`
	TotalMs int64 `json:"total_ms"`

	// Commands are the command keys by total time, Tools the same
	// rolled up per tool, Usage a row per tool per day, and
	// CommandUsage a row per command key per day. An atkins older than
	// the key reports no CommandUsage.
	Commands     []Command      `json:"commands"`
	Tools        []Tool         `json:"tools"`
	Usage        []Usage        `json:"usage"`
	CommandUsage []CommandUsage `json:"command_usage"`

	// Samples are the commands of the window one by one, oldest first,
	// which a chart buckets finer than a day. An atkins older than
	// `--stats --all` reports none.
	Samples []Sample `json:"samples,omitempty"`
}

// Command is one command key and what it has cost. The key is the
// command's words up to its first flag, at most three and no path or
// value: `go test -v ./pkg` counts as `go test`, and `docker compose
// up -d` as `docker compose up`.
type Command struct {
	Command   string     `json:"command"`
	Tool      string     `json:"tool"`
	Runs      int64      `json:"runs"`
	TotalMs   int64      `json:"total_ms"`
	AverageMs int64      `json:"average_ms"`
	MaxMs     int64      `json:"max_ms"`
	Failures  int64      `json:"failures"`
	LastRunAt *time.Time `json:"last_run_at,omitempty"`
}

// Tool is one tool and what it has cost across every form of it.
type Tool struct {
	Tool     string `json:"tool"`
	Commands int64  `json:"commands"`
	Runs     int64  `json:"runs"`
	TotalMs  int64  `json:"total_ms"`
	Failures int64  `json:"failures"`
}

// Usage is one tool on one day, the series behind a chart of tool use
// over time. A day a tool did not run has no row.
type Usage struct {
	Date    string `json:"date"`
	Tool    string `json:"tool"`
	Runs    int64  `json:"runs"`
	TotalMs int64  `json:"total_ms"`
}

// Empty reports a scope with nothing recorded in it.
func (r *Scope) Empty() bool {
	return r == nil || len(r.Usage) == 0
}

// Since returns the usage rows of the last days, counted back from the report's last day, oldest first.
func (r *Scope) Since(days int) []Usage {
	if r == nil || days <= 0 {
		return nil
	}

	last := r.lastDate()
	if last.IsZero() {
		return nil
	}
	first := last.AddDate(0, 0, -(days - 1)).Format(time.DateOnly)

	var rows []Usage
	for _, row := range r.Usage {
		if row.Date >= first {
			rows = append(rows, row)
		}
	}
	return rows
}

// ToolsSince returns the tools of the last days, rolled up from the usage rows, most time taken first.
func (r *Scope) ToolsSince(days int) []Tool {
	totals := map[string]*Tool{}
	for _, row := range r.Since(days) {
		tool, ok := totals[row.Tool]
		if !ok {
			tool = &Tool{Tool: row.Tool}
			totals[row.Tool] = tool
		}
		tool.Runs += row.Runs
		tool.TotalMs += row.TotalMs
	}

	tools := make([]Tool, 0, len(totals))
	for _, tool := range totals {
		tools = append(tools, *tool)
	}
	sort.Slice(tools, func(i, j int) bool {
		if tools[i].TotalMs != tools[j].TotalMs {
			return tools[i].TotalMs > tools[j].TotalMs
		}
		return tools[i].Tool < tools[j].Tool
	})

	return tools
}

// Daily returns the total time per day of the last days, oldest first, with a zero for a day nothing ran.
func (r *Scope) Daily(days int) []Day {
	if r == nil || days <= 0 {
		return nil
	}

	last := r.lastDate()
	if last.IsZero() {
		return nil
	}

	totals := map[string]*Day{}
	for _, row := range r.Since(days) {
		day, ok := totals[row.Date]
		if !ok {
			day = &Day{Date: row.Date}
			totals[row.Date] = day
		}
		day.Runs += row.Runs
		day.TotalMs += row.TotalMs
	}

	out := make([]Day, 0, days)
	for at := last.AddDate(0, 0, -(days - 1)); !at.After(last); at = at.AddDate(0, 0, 1) {
		date := at.Format(time.DateOnly)
		if day, ok := totals[date]; ok {
			out = append(out, *day)
			continue
		}
		out = append(out, Day{Date: date})
	}

	return out
}

// Day is one day of the window: everything that ran in it.
type Day struct {
	Date    string
	Runs    int64
	TotalMs int64
}

// lastDate is the day the report ends on, the day atkins made it.
func (r *Scope) lastDate() time.Time {
	at, err := time.Parse(time.DateOnly, strings.TrimSpace(r.To))
	if err != nil {
		return time.Time{}
	}
	return at
}
