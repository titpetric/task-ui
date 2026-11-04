package phpfront

import (
	"reflect"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// The keys the tooling report indexes by. The helpers are convert_test.go's.

// A scope is one repository's window: what atkins recorded, the commands
// summed, the two daily series and the commands one by one. The page draws
// four charts off the one answer, so all four cross together.
func TestToolingScopes(t *testing.T) {
	t.Parallel()

	out := toolingScopes([]ToolingScope{{
		Name: "task-ui", Repository: "task-ui", Database: "/data/atkins.db",
		Days: 30, From: "2026-10-04", To: "2026-11-02", Runs: 3, TotalMs: 4500,
		Commands:     []ToolingCommand{{Command: "go test", Tool: "go"}},
		Usage:        []ToolingUsage{{Date: "2026-11-02", Tool: "go", Runs: 3, TotalMs: 4500}},
		CommandUsage: []ToolingUsage{{Date: "2026-11-02", Command: "go test", Tool: "go", Runs: 3}},
		Samples:      []ToolingSample{{Command: "go test"}},
	}})
	assert.Equal(t, 1, out.Len())

	scope := at(t, out, 0)
	assert.Equal(t, "task-ui", text(t, scope, "name"))
	assert.Equal(t, "/data/atkins.db", text(t, scope, "database"))
	assert.Equal(t, 30, value[int](t, scope, "days"))
	assert.Equal(t, "2026-10-04", text(t, scope, "from"))
	assert.Equal(t, "2026-11-02", text(t, scope, "to"))

	for _, key := range []string{"commands", "usage", "command_usage", "samples"} {
		assert.Equal(t, 1, list(t, scope, key).Len(), key)
	}

	// The two series are the same shape under different keys: one by tool
	// and one by command, so a chart of either reads the same row.
	assert.Equal(t, "", text(t, at(t, list(t, scope, "usage"), 0), "command"))
	assert.Equal(t, "go test", text(t, at(t, list(t, scope, "command_usage"), 0), "command"))

	// A repository atkins recorded nothing for still answers every key, so
	// the page draws its heading and four empty charts.
	empty := at(t, toolingScopes([]ToolingScope{{Name: "task-ui"}}), 0)
	for _, key := range []string{"commands", "usage", "command_usage", "samples"} {
		assert.Equal(t, 0, list(t, empty, key).Len(), key)
	}

	assert.Equal(t, 0, toolingScopes(nil).Len())
}

// A command key crosses with the arithmetic already done: the total, the mean,
// the worst and the failures, because the table sorts on each of them and a
// template that computed a mean would compute it per sort.
func TestToolingCommands(t *testing.T) {
	t.Parallel()

	out := toolingCommands([]ToolingCommand{{
		Command: "go test", Tool: "go", Runs: 3, TotalMs: 4500,
		AverageMs: 1500, MaxMs: 2500, Failures: 1,
		LastRunAt: "2026-11-02 10:32:12.414 +0000 UTC",
	}})
	assert.Equal(t, 1, out.Len())

	row := at(t, out, 0)
	for i := range reflect.TypeFor[ToolingCommand]().NumField() {
		key := reflect.TypeFor[ToolingCommand]().Field(i).Tag.Get("json")
		_, ok := row.Get(key)
		assert.True(t, ok, key)
	}

	assert.Equal(t, "go test", text(t, row, "command"))
	assert.Equal(t, "go", text(t, row, "tool"), "the first word, which the by-tool chart groups on")
	assert.Equal(t, int64(4500), number(t, row, "total_ms"))
	assert.Equal(t, int64(1500), number(t, row, "average_ms"))
	assert.Equal(t, int64(2500), number(t, row, "max_ms"))
	assert.Equal(t, int64(1), number(t, row, "failures"))
	assert.Equal(t, "2026-11-02 10:32:12.414 +0000 UTC", text(t, row, "last_run_at"))

	assert.Equal(t, 0, toolingCommands(nil).Len())
}

// The daily series is one shape for both charts: a row with a command is the
// by-command one and a row without is the by-tool one, so the date and the
// two measures are read the same way either side.
func TestToolingUsage(t *testing.T) {
	t.Parallel()

	out := toolingUsage([]ToolingUsage{
		{Date: "2026-11-01", Tool: "go", Runs: 2, TotalMs: 2000},
		{Date: "2026-11-02", Command: "go test", Tool: "go", Runs: 3, TotalMs: 4500},
	})
	assert.Equal(t, 2, out.Len())

	assert.Equal(t, "2026-11-01", text(t, at(t, out, 0), "date"))
	assert.Equal(t, int64(2), number(t, at(t, out, 0), "runs"))
	assert.Equal(t, int64(2000), number(t, at(t, out, 0), "total_ms"))
	assert.Equal(t, "go test", text(t, at(t, out, 1), "command"))

	assert.Equal(t, 0, toolingUsage(nil).Len())
}

// The same reflection over a sample of the report, which is read by index five
// thousand times a page: a field the converter does not list is a chart
// drawing nothing, and at_unix is there to be read instead of parsed.
func TestToolingSamples(t *testing.T) {
	t.Parallel()

	out := toolingSamples([]ToolingSample{{
		At: "2026-11-02 10:32:12.414 +0000 UTC", AtUnix: 1793961132,
		Command: "go test", Tool: "go", Job: "test", DurationMs: 2500, ExitCode: 1, Failed: true,
	}})
	assert.Equal(t, 1, out.Len())

	row := at(t, out, 0)
	for i := range reflect.TypeFor[ToolingSample]().NumField() {
		key := reflect.TypeFor[ToolingSample]().Field(i).Tag.Get("json")
		_, ok := row.Get(key)
		assert.True(t, ok, key)
	}

	assert.Equal(t, "2026-11-02 10:32:12.414 +0000 UTC", text(t, row, "at"))
	assert.Equal(t, int64(1793961132), number(t, row, "at_unix"))
	assert.Equal(t, "test", text(t, row, "job"), "the job of the run that spent it, where one did")
	assert.Equal(t, int64(1), number(t, row, "exit_code"))
	assert.True(t, flag(t, row, "failed"))

	assert.Equal(t, 0, toolingSamples(nil).Len())
}
