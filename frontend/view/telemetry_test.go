package view_test

import (
	"testing"
	"time"

	"github.com/titpetric/oida"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The two list fields read as the textareas that edit them spell them,
// one entry a line, and come back the same way.
func TestTelemetry(t *testing.T) {
	t.Parallel()
	opts := oida.NewOptions("task-ui")
	opts.IgnorePaths = []string{"/healthz", "/metrics"}
	opts.AllowedNetworks = []string{"10.0.0.0/8"}
	shown := &view.Telemetry{Options: opts, Path: oida.DefaultPath, Admin: true}

	assert.Equal(t, "/healthz\n/metrics", shown.IgnorePaths())
	assert.Equal(t, "10.0.0.0/8", shown.AllowedNetworks())
	assert.Equal(t, []string{"/healthz", "/metrics"}, model.Lines(shown.IgnorePaths()))

	empty := &view.Telemetry{Options: oida.Options{}}
	assert.Equal(t, "", empty.IgnorePaths())
	assert.Equal(t, "", empty.AllowedNetworks())
}

// The shortest debug span reads as a duration, which is what the field posts
// back: a number of nanoseconds is not a unit anybody typing in the box means.
func TestTelemetryDebugSpanMinDuration(t *testing.T) {
	t.Parallel()
	opts := oida.NewOptions("task-ui")
	opts.DebugSpanMinDuration = time.Millisecond
	shown := &view.Telemetry{Options: opts}

	assert.Equal(t, "1ms", shown.DebugSpanMinDuration())
	assert.Equal(t, "0s", (&view.Telemetry{}).DebugSpanMinDuration())
}

// The page shows every setting, and the one that takes effect on its own
// says so beside the ones that rebuild the recorder.
func TestTelemetryPage(t *testing.T) {
	t.Parallel()
	opts := oida.NewOptions("task-ui")
	opts.Enabled = true
	opts.SampleRate = 25
	opts.RingBufferSize = 64

	p := admin(t)
	p.Path = "/settings/telemetry"
	p.Step = view.StepTelemetry
	p.Telemetry = &view.Telemetry{Options: opts, Path: "/debug/oida", Admin: true}
	body := render(t, "telemetry", p)

	assert.Contains(t, body, `data-api="/api/telemetry"`)
	assert.Contains(t, body, `name="enabled"`)
	assert.Contains(t, body, `value="25"`)
	assert.Contains(t, body, `value="64"`)
	assert.Contains(t, body, `name="debug_span_min_duration"`)
	assert.Contains(t, body, `name="ignore_paths"`)
	assert.Contains(t, body, `name="allowed_networks"`)
	assert.Contains(t, body, `href="/debug/oida"`)
	assert.Contains(t, body, "takes a restart")

	// A viewer reads the page and saves nothing.
	p.Telemetry.Admin = false
	body = render(t, "telemetry", p)
	assert.NotContains(t, body, `data-api="/api/telemetry"`)
	assert.Contains(t, body, "An admin changes the settings.")
	assert.Contains(t, body, "disabled")
}
