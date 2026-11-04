package app

import (
	"context"

	"github.com/titpetric/task-ui/phpfront"
	"github.com/titpetric/task-ui/tooling"
)

// Tooling is what atkins has recorded for the tree, the scopes of its report
// with the project's first.
//
// It reports commands and not runs, so it says nothing about who ran what and
// needs no permission of its own beyond the page it sits on: a reader of the
// runs page may already see that the tree runs go test, because the runs are
// there.
//
// Empty without a prober, without a root, and on a tree nobody has run a job
// in through atkins, which is the usual answer.
func (h *FrontHost) Tooling(context.Context) []phpfront.ToolingScope {
	if h.Tools == nil || h.Root == "" {
		return nil
	}
	found := h.Tools.Probe(h.Root).Scopes()
	out := make([]phpfront.ToolingScope, 0, len(found))
	for _, scope := range found {
		out = append(out, frontToolingScope(scope))
	}
	return out
}

// frontToolingScope is one database of the report.
func frontToolingScope(scope *tooling.Scope) phpfront.ToolingScope {
	out := phpfront.ToolingScope{
		Name:       scope.Name,
		Repository: scope.Repository,
		Database:   scope.Database,
		Days:       scope.Days,
		From:       scope.From,
		To:         scope.To,
		Runs:       scope.Runs,
		TotalMs:    scope.TotalMs,
	}
	for _, command := range scope.Commands {
		each := phpfront.ToolingCommand{
			Command:   command.Command,
			Tool:      command.Tool,
			Runs:      command.Runs,
			TotalMs:   command.TotalMs,
			AverageMs: command.AverageMs,
			MaxMs:     command.MaxMs,
			Failures:  command.Failures,
		}
		if command.LastRunAt != nil {
			each.LastRunAt = command.LastRunAt.String()
		}
		out.Commands = append(out.Commands, each)
	}
	for _, row := range scope.Usage {
		out.Usage = append(out.Usage, phpfront.ToolingUsage{Date: row.Date, Tool: row.Tool, Runs: row.Runs, TotalMs: row.TotalMs})
	}
	for _, row := range scope.CommandUsage {
		out.CommandUsage = append(out.CommandUsage, phpfront.ToolingUsage{
			Date: row.Date, Command: row.Command, Tool: row.Tool, Runs: row.Runs, TotalMs: row.TotalMs,
		})
	}
	for _, sample := range scope.Samples {
		each := phpfront.ToolingSample{
			At:         sample.At.String(),
			Command:    sample.Command,
			Tool:       sample.Tool,
			Job:        sample.Job,
			DurationMs: sample.DurationMs,
			ExitCode:   sample.ExitCode,
			Failed:     sample.Failed,
		}
		// A sample with no moment is on no chart, and 0 is how PHP spells
		// that: the unix time of a zero time is a year the charts would
		// have to reach back to to leave it out.
		if !sample.At.IsZero() {
			each.AtUnix = sample.At.Unix()
		}
		out.Samples = append(out.Samples, each)
	}
	return out
}
