// Package tooling asks atkins what a checkout has been running.
//
// `atkins --stats --json` reports the commands a run executed, rolled up
// per tool and per day, over the window atkins keeps. task-ui reads it
// to answer a question its own run history cannot: a run of a job is one
// row here, and the twenty commands inside it are twenty rows there, so
// the time a repository spends on go, npm or docker is atkins' to
// report and not task-ui's to guess.
//
// Nothing in this package runs a command of its own. It execs atkins,
// decodes what it prints, and keeps the answer for a while.
package tooling

import (
	"sort"
	"strings"
	"time"
)

// Report is what atkins says about a directory.
type Report struct {
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

	// Commands are the normalized commands by total time, Tools the
	// same rolled up per tool, and Usage a row per tool per day.
	Commands []Command `json:"commands"`
	Tools    []Tool    `json:"tools"`
	Usage    []Usage   `json:"usage"`
}

// Command is one normalized command and what it has cost: `go test -v`
// counts apart from `go test`, and `go test ./pkg` counts with it.
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

// Usage is one tool on one day: the series a chart of tool use over
// time is drawn from. A day a tool did not run has no row.
type Usage struct {
	Date    string `json:"date"`
	Tool    string `json:"tool"`
	Runs    int64  `json:"runs"`
	TotalMs int64  `json:"total_ms"`
}

// Empty reports a report with nothing recorded in it.
func (r *Report) Empty() bool {
	return r == nil || len(r.Usage) == 0
}

// Since returns the usage rows of the last days, counted back from the
// report's last day, oldest first.
//
// The window is a count of days and not a cutoff time, because the rows
// are dates and the question is "the last seven days", which a reader
// asks of a calendar and not of a clock.
func (r *Report) Since(days int) []Usage {
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

// ToolsSince returns the tools of the last days by the time they have
// taken, the most expensive first.
//
// The report's own Tools cover its whole window, so a shorter one is
// rolled up from the usage rows rather than taken from them.
func (r *Report) ToolsSince(days int) []Tool {
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

// Daily returns the total time per day of the last days, oldest first,
// with a zero for a day nothing ran.
//
// The empty days are filled because a run of days reads as time: a chart
// that leaves them out draws a busy fortnight and a quiet one the same
// width.
func (r *Report) Daily(days int) []Day {
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

// lastDate is the day the report ends on, which is the day it was made.
func (r *Report) lastDate() time.Time {
	at, err := time.Parse(time.DateOnly, strings.TrimSpace(r.To))
	if err != nil {
		return time.Time{}
	}
	return at
}
