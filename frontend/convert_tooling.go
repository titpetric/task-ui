package phpfront

import (
	"github.com/titpetric/phpscript/model"
)

// toolingScopes is what atkins recorded as a PHP array.
func toolingScopes(found []ToolingScope) *model.Array {
	out := model.NewArraySize(len(found))
	for _, scope := range found {
		out.Append(phpMap(
			"name", scope.Name,
			"repository", scope.Repository,
			"database", scope.Database,
			"days", scope.Days,
			"from", scope.From,
			"to", scope.To,
			"runs", scope.Runs,
			"total_ms", scope.TotalMs,
			"commands", toolingCommands(scope.Commands),
			"usage", toolingUsage(scope.Usage),
			"command_usage", toolingUsage(scope.CommandUsage),
			"samples", toolingSamples(scope.Samples),
		))
	}
	return out
}

// toolingCommands is a scope's command keys as a PHP array.
func toolingCommands(found []ToolingCommand) *model.Array {
	out := model.NewArraySize(len(found))
	for _, command := range found {
		out.Append(phpMap(
			"command", command.Command,
			"tool", command.Tool,
			"runs", command.Runs,
			"total_ms", command.TotalMs,
			"average_ms", command.AverageMs,
			"max_ms", command.MaxMs,
			"failures", command.Failures,
			"last_run_at", command.LastRunAt,
		))
	}
	return out
}

// toolingUsage is a daily series as a PHP array.
func toolingUsage(found []ToolingUsage) *model.Array {
	out := model.NewArraySize(len(found))
	for _, row := range found {
		out.Append(phpMap(
			"date", row.Date,
			"command", row.Command,
			"tool", row.Tool,
			"runs", row.Runs,
			"total_ms", row.TotalMs,
		))
	}
	return out
}

// toolingSamples is the commands of a window one by one as a PHP array.
func toolingSamples(found []ToolingSample) *model.Array {
	out := model.NewArraySize(len(found))
	for _, sample := range found {
		out.Append(phpMap(
			"at", sample.At,
			"at_unix", sample.AtUnix,
			"command", sample.Command,
			"tool", sample.Tool,
			"job", sample.Job,
			"duration_ms", sample.DurationMs,
			"exit_code", sample.ExitCode,
			"failed", sample.Failed,
		))
	}
	return out
}
