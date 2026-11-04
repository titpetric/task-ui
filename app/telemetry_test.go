package app_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
)

// An instance that stored no telemetry options drops a span marked as detail
// that ran under a millisecond.
//
// The library's own minimum is a microsecond, which is the floor under which a
// span measures the cost of measuring. The producer here is the PHP
// interpreter, which marks a span per frame, so the instance names a figure a
// reader of the trace means: a frame faster than a millisecond is not where
// the page spent its time.
func TestTelemetryDefaults(t *testing.T) {
	t.Parallel()
	telemetry := &app.Telemetry{}
	opts := telemetry.Defaults()

	assert.Equal(t, time.Millisecond, opts.DebugSpanMinDuration)
	assert.NotNil(t, opts.Authorize, "the dashboard is behind an admin")
	assert.NotNil(t, opts.RouteFunc, "and the statistics group by the routed pattern")
}
