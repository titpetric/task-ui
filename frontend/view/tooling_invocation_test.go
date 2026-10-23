package view_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/tooling"
)

func TestToolingInvocation(t *testing.T) {
	t.Parallel()

	var invocation view.ToolingInvocation = view.NewTooling(&tooling.Report{Project: sampleScope()}, toolingNow).Scopes[0].Latest[0]
	assert.Equal(t, "go test", invocation.Command)
	assert.False(t, invocation.Failed)
}

// sampleScope is a month of history listed one command at a time: go
// test every hour of the last day, seven other commands once each this
// evening, and a build a fortnight back.
func sampleScope() *tooling.Scope {
	scope := toolingScope()
	scope.Samples = nil
	for hour := range 24 {
		scope.Samples = append(scope.Samples, tooling.Sample{
			At: toolingNow.Add(-time.Duration(23-hour) * time.Hour), Command: "go test", Tool: "go", DurationMs: 1000,
		})
	}
	for index, command := range []string{"docker build", "docker compose up", "git clone", "npm run build", "go vet", "make", "curl"} {
		scope.Samples = append(scope.Samples, tooling.Sample{
			At: toolingNow.Add(-time.Duration(30-index) * time.Minute), Command: command, Tool: strings.Fields(command)[0],
			DurationMs: int64(900 - 100*index), Job: "build", ExitCode: int64(index % 2),
		})
	}
	scope.Samples = append(scope.Samples, tooling.Sample{At: toolingNow.AddDate(0, 0, -14), Command: "docker build", Tool: "docker", DurationMs: 60_000})
	// atkins lists them oldest first.
	sort.Slice(scope.Samples, func(i, j int) bool { return scope.Samples[i].At.Before(scope.Samples[j].At) })
	return scope
}

// The newest commands, newest first, at most twenty, in their chart
// colours.
func TestToolingLatest(t *testing.T) {
	t.Parallel()

	scope := view.NewTooling(&tooling.Report{Project: sampleScope()}, toolingNow).Scopes[0]
	assert.Len(t, scope.Latest, 20)

	newest := scope.Latest[0]
	assert.Equal(t, "go test", newest.Command)
	assert.Equal(t, "18:00", newest.When)
	assert.Equal(t, time.Second, newest.Took)
	assert.NotEqual(t, "s-other", newest.Color)

	curl := scope.Latest[1]
	assert.Equal(t, "curl", curl.Command)
	assert.Equal(t, "build", curl.Job)
	assert.Equal(t, "s-other", curl.Color)

	make := scope.Latest[2]
	assert.True(t, make.Failed)
	assert.Equal(t, int64(1), make.Exit)

	assert.Equal(t, "06:00", scope.Latest[19].When)

	// A command of an earlier day carries its weekday.
	report := &tooling.Report{Project: sampleScope()}
	report.Project.Samples = report.Project.Samples[:1]
	assert.Equal(t, "Fri 18:00", view.NewTooling(report, toolingNow.AddDate(0, 0, -7)).Scopes[0].Latest[0].When)
}

// The panel of samples renders the three charts, the week on its own
// row, and the newest commands.
func TestToolingRendersSamples(t *testing.T) {
	v, err := view.New()
	assert.NoError(t, err)

	var out strings.Builder
	page := &view.Page{Title: "Runs", Tooling: view.NewTooling(&tooling.Report{Project: sampleScope()}, toolingNow)}
	assert.NoError(t, v.Render(&out, "runs", page))

	html := out.String()
	for _, want := range []string{
		"Last 24 hours", "busiest hour", "busiest 3 hours",
		`class="tooling-chart wide"`, `class="tooling-bars fine" style="--n: 56; --k: 6"`,
		`aria-label="Time spent per hour by command, last 24 hours"`,
		"Latest commands", `<span class="bad" title="exit 1">failed</span>`,
		fmt.Sprintf(`<time datetime="%s"`, view.Stamp(toolingNow)),
	} {
		assert.Contains(t, html, want)
	}
}
