//go:build integration

package api_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/oida"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

// stubTelemetry is the settings the API writes, without a tracer: it
// holds the last save and refuses what the real one refuses.
type stubTelemetry struct {
	opts oida.Options
}

func (s *stubTelemetry) Load(_ context.Context) (oida.Options, error) { return s.opts, nil }

func (s *stubTelemetry) Save(_ context.Context, opts oida.Options) error {
	if err := opts.Validate(); err != nil {
		return err
	}
	s.opts = opts
	return nil
}

// withTelemetry puts the settings on the fixture and rebuilds the routes.
func withTelemetry(t *testing.T, f *fixture) *stubTelemetry {
	t.Helper()
	stub := &stubTelemetry{opts: oida.NewOptions("task-ui")}
	f.opts.Telemetry = stub
	f.reload(t)
	return stub
}

// A request names the settings it changes and the rest keep what they
// have, so the page posts the whole form and an agent posts the switch.
func TestTelemetry(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	stub := withTelemetry(t, f)

	list := f.call(t, asAdmin, http.MethodGet, "/telemetry", nil)
	assert.Equal(t, http.StatusOK, list.Status)
	assert.Equal(t, "task-ui", list.str("service_name"))
	assert.Equal(t, false, list.get("enabled"))

	saved := f.call(t, asAdmin, http.MethodPut, "/telemetry", map[string]any{"enabled": true})
	assert.Equal(t, http.StatusOK, saved.Status)
	assert.Equal(t, true, saved.get("enabled"))
	assert.Equal(t, "task-ui", saved.str("service_name"), "a setting the request left out keeps its value")
	assert.True(t, stub.opts.Enabled)
	assert.Equal(t, 200, stub.opts.RingBufferSize)

	// The two list fields arrive as the textareas spell them, and a blank
	// line is not an entry: an empty path would match every path.
	saved = f.call(t, asAdmin, http.MethodPut, "/telemetry", map[string]any{
		"ignore_paths":     "/healthz\n\n  /metrics  \n",
		"allowed_networks": "10.0.0.0/8\n",
		"sample_rate":      25,
		"ring_buffer_size": 64,
	})
	assert.Equal(t, http.StatusOK, saved.Status)
	assert.Equal(t, []string{"/healthz", "/metrics"}, stub.opts.IgnorePaths)
	assert.Equal(t, []string{"10.0.0.0/8"}, stub.opts.AllowedNetworks)
	assert.Equal(t, float64(25), stub.opts.SampleRate)
	assert.Equal(t, 64, stub.opts.RingBufferSize)
	assert.True(t, stub.opts.Enabled, "the switch stayed on")
}

// Only an admin changes the settings, and a value out of range is the
// caller's mistake rather than the instance's.
func TestTelemetryRefuses(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	stub := withTelemetry(t, f)

	for _, user := range []string{asOps, asViewer} {
		res := f.call(t, user, http.MethodPut, "/telemetry", map[string]any{"enabled": true})
		assert.Equal(t, http.StatusForbidden, res.Status, user)
	}
	assert.False(t, stub.opts.Enabled)

	res := f.call(t, asAdmin, http.MethodPut, "/telemetry", map[string]any{"sample_rate": 400})
	assert.Equal(t, http.StatusBadRequest, res.Status)
	assert.Contains(t, res.str("error"), "sample rate")

	res = f.call(t, asAdmin, http.MethodPut, "/telemetry", map[string]any{"path": "debug/oida"})
	assert.Equal(t, http.StatusBadRequest, res.Status)
	assert.Equal(t, oida.DefaultPath, stub.opts.Path, "a refused save changes nothing")
}

// An instance with no telemetry wired says so rather than pretending.
func TestTelemetryNone(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	res := f.call(t, asAdmin, http.MethodGet, "/telemetry", nil)
	assert.Equal(t, http.StatusNotFound, res.Status)

	res = f.call(t, asAdmin, http.MethodPut, "/telemetry", map[string]any{"enabled": true})
	assert.Equal(t, http.StatusNotFound, res.Status)
	assert.Contains(t, res.str("error"), "isn't set up")
}

// The request type names the yaml fields of oida.Options, so the form,
// the stored row and the library spell a setting the same way.
func TestTelemetryRequest(t *testing.T) {
	t.Parallel()
	opts := oida.NewOptions("task-ui")
	enabled, rate := true, 40.0
	req := api.TelemetryRequest{Enabled: &enabled, SampleRate: &rate}

	before := opts.RingBufferSize
	req.Apply(&opts)
	assert.True(t, opts.Enabled)
	assert.Equal(t, 40.0, opts.SampleRate)
	assert.Equal(t, before, opts.RingBufferSize, "a field the request left out is untouched")
}
