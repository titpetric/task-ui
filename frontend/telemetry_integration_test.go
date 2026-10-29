//go:build integration

package frontend_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/oida"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/permission"
)

// fakeTelemetry is the settings the page reads, without a tracer.
type fakeTelemetry struct {
	opts oida.Options
	path string
	err  error
}

func (f *fakeTelemetry) Load(_ context.Context) (oida.Options, error) { return f.opts, f.err }
func (f *fakeTelemetry) Path() string                                 { return f.path }

// The page is the one form that configures the recorder, with the switch
// that takes effect on its own and the path that takes a restart; an
// admin saves, a user who may see the settings reads it, and a viewer
// may not open it.
func TestTelemetrySettings(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()

	opts := oida.NewOptions("task-ui")
	opts.Enabled, opts.SampleRate, opts.RingBufferSize = true, 25, 64
	opts.IgnorePaths = []string{"/healthz", "/metrics"}
	f.opts.Telemetry = &fakeTelemetry{opts: opts, path: "/debug/oida"}
	f.reload(t)

	body := f.page(t, asAdmin, "/settings/telemetry")
	assert.Contains(t, body, "<h2>Telemetry</h2>")
	assert.Contains(t, body, `data-api="/api/telemetry"`)
	assert.Contains(t, body, `name="enabled"`)
	assert.Contains(t, body, `value="25"`)
	assert.Contains(t, body, `value="64"`)
	assert.Contains(t, body, "/healthz\n/metrics", "the list reads as the textarea spells it")
	assert.Contains(t, body, `href="/debug/oida"`)
	assert.Contains(t, body, "takes a restart", "the path is read once, at startup")
	assert.Contains(t, body, "empties the traces it holds", "what a save costs")

	assert.NoError(t, f.opts.Users.Policies().Put(ctx, permission.KindUser, asOps, "secret/view", permission.Allow))
	body = f.page(t, asOps, "/settings/telemetry")
	assert.NotContains(t, body, `data-api="/api/telemetry"`)
	assert.Contains(t, body, "An admin changes the settings.")

	assert.Equal(t, http.StatusForbidden, f.get(t, asViewer, "/settings/telemetry").Status)
}

// An instance with no telemetry wired shows the defaults, read only,
// rather than an empty page.
func TestTelemetrySettingsNone(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/settings/telemetry")
	assert.Contains(t, body, "<h2>Telemetry</h2>")
	assert.NotContains(t, body, `data-api="/api/telemetry"`)
	assert.Contains(t, body, `href="`+oida.DefaultPath+`"`)
	assert.Contains(t, body, "An admin changes the settings.")
}

// The step is in the settings trail, after Data retention.
func TestTelemetrySettingsStep(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/settings")
	assert.Contains(t, body, `href="/settings/telemetry"`)
	assert.Contains(t, body, "What the instance records of its own requests, and who reads the traces.")

	body = f.page(t, asAdmin, "/settings/telemetry")
	assert.Contains(t, body, `href="/settings/retention"`, "the trail links the step before it")
}
