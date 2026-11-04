package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/titpetric/oida"

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
	telemetry := &Telemetry{}
	opts := telemetry.Defaults()

	assert.Equal(t, time.Millisecond, opts.DebugSpanMinDuration)
	assert.NotNil(t, opts.Authorize, "the dashboard is behind an admin")
	assert.NotNil(t, opts.RouteFunc, "and the statistics group by the routed pattern")
}

// An instance that built no tracer has no debug front end. The route is there
// either way, because the router is built once at start and a save changes the
// setting under it, so what the handler answers is a page that is not there.
func TestTelemetry_ServeHTTP(t *testing.T) {
	t.Parallel()
	telemetry := &Telemetry{}
	request := httptest.NewRequest(http.MethodGet, "/debug/oida/", nil)

	assert.ErrorIs(t, telemetry.serveHTTP(httptest.NewRecorder(), request), errTelemetryOff)

	recorder := httptest.NewRecorder()
	telemetry.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

// Tracer is nil until one is built, which is what every reader of the pointer
// checks for, and Path falls back to the library's mount so the chrome has a
// link to draw.
func TestTelemetry_Tracer(t *testing.T) {
	t.Parallel()

	var absent *Telemetry
	assert.Nil(t, absent.Tracer(), "a nil telemetry is an instance without one")

	telemetry := &Telemetry{}
	assert.Nil(t, telemetry.Tracer())
	assert.Equal(t, oida.DefaultPath, telemetry.Path())
}
