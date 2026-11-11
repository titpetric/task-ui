package phpfront_test

import (
	"context"

	"github.com/titpetric/task-ui/phpfront"
)

// What the fixture answers of the tooling report: one scope of what atkins
// recorded, grouped the way host_tooling.go declares the shapes.

func (h *host) Tooling(context.Context) []phpfront.ToolingScope {
	return []phpfront.ToolingScope{{
		Name: "task-ui", Repository: "task-ui", Database: "/data/atkins.db",
		Days: 30, From: "2026-10-04", To: "2026-11-02", Runs: 3, TotalMs: 4500,
		Commands: []phpfront.ToolingCommand{{
			Command: "go test", Tool: "go", Runs: 3, TotalMs: 4500,
			AverageMs: 1500, MaxMs: 2500, Failures: 1,
			LastRunAt: "2026-11-02 10:32:12.414 +0000 UTC",
		}},
		Usage:        []phpfront.ToolingUsage{{Date: "2026-11-02", Tool: "go", Runs: 3, TotalMs: 4500}},
		CommandUsage: []phpfront.ToolingUsage{{Date: "2026-11-02", Command: "go test", Tool: "go", Runs: 3, TotalMs: 4500}},
		Samples: []phpfront.ToolingSample{{
			At: "2026-11-02 10:32:12.414 +0000 UTC", Command: "go test", Tool: "go",
			Job: "test", DurationMs: 2500, ExitCode: 1, Failed: true,
		}},
	}}
}
