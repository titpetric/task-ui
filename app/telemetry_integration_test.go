//go:build integration

package app_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/oida"
	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
	_ "modernc.org/sqlite"
)

// newSettings opens a database of its own and returns its setting storage.
func newSettings(t *testing.T) *storage.SettingStorage {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })
	return storage.NewSettingStorage(db)
}

// An instance with nothing stored starts on oida's defaults, and a save
// stores the options and applies them.
func TestTelemetry(t *testing.T) {
	t.Parallel()
	settings := newSettings(t)
	ctx := context.Background()

	telemetry, err := app.NewTelemetry(ctx, settings, "task-ui-test")
	assert.NoError(t, err)
	assert.NotNil(t, telemetry.Tracer())

	opts, err := telemetry.Load(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "task-ui-test", opts.ServiceName)
	assert.Equal(t, oida.DefaultPath, opts.Path)
	assert.Equal(t, 200, opts.RingBufferSize)

	opts.Enabled = true
	opts.SampleRate = 50
	opts.RingBufferSize = 64
	assert.NoError(t, telemetry.Save(ctx, opts))

	// The stored row is what a later start reads.
	stored, err := telemetry.Load(ctx)
	assert.NoError(t, err)
	assert.True(t, stored.Enabled)
	assert.Equal(t, float64(50), stored.SampleRate)
	assert.Equal(t, 64, stored.RingBufferSize)
	assert.False(t, stored.ReadEnv, "a stored row stops the environment deciding")

	values, err := settings.Values(ctx)
	assert.NoError(t, err)
	assert.Contains(t, values[app.TelemetrySetting], "ring_buffer_size: 64")
	assert.NotContains(t, values[app.TelemetrySetting], "route_func", "a callback is not a setting")
}

// Enabled alone keeps the tracer and the traces it holds; any other
// change builds a new one, which starts an empty ring buffer.
func TestTelemetryApply(t *testing.T) {
	t.Parallel()
	telemetry, err := app.NewTelemetry(context.Background(), newSettings(t), "task-ui-test")
	assert.NoError(t, err)

	first := telemetry.Tracer()
	assert.False(t, first.Enabled())

	opts := telemetry.Defaults()
	opts.Enabled = true
	assert.NoError(t, telemetry.Apply(opts))
	assert.Equal(t, first, telemetry.Tracer(), "Enabled alone keeps the tracer in place")
	assert.True(t, telemetry.Tracer().Enabled())

	opts.RingBufferSize = 32
	assert.NoError(t, telemetry.Apply(opts))
	assert.NotEqual(t, first, telemetry.Tracer(), "any other change swaps the tracer")
	assert.True(t, telemetry.Tracer().Enabled())
	assert.Equal(t, 32, telemetry.Tracer().Options().RingBufferSize)
}

// A sample rate outside its range is refused, and nothing is stored.
func TestTelemetryRefuses(t *testing.T) {
	t.Parallel()
	settings := newSettings(t)
	ctx := context.Background()
	telemetry, err := app.NewTelemetry(ctx, settings, "task-ui-test")
	assert.NoError(t, err)

	opts := telemetry.Defaults()
	opts.SampleRate = 400
	assert.ErrorIs(t, telemetry.Save(ctx, opts), oida.ErrInvalidSampleRate)

	opts = telemetry.Defaults()
	opts.Path = "debug/oida"
	assert.ErrorIs(t, telemetry.Save(ctx, opts), oida.ErrInvalidPath)

	values, err := settings.Values(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "", values[app.TelemetrySetting], "a refused save stores nothing")
}

// A stored row that does not parse comes back as an error over the
// defaults, so a hand-edited setting cannot leave the instance with a
// tracer nobody configured.
func TestTelemetryBadSetting(t *testing.T) {
	t.Parallel()
	settings := newSettings(t)
	ctx := context.Background()
	assert.NoError(t, settings.Set(ctx, map[string]string{app.TelemetrySetting: "sample_rate: [not a number]"}))

	telemetry, err := app.NewTelemetry(ctx, settings, "task-ui-test")
	assert.Error(t, err)
	assert.Nil(t, telemetry)
}

// The middleware passes a request through whether or not a tracer is
// recording, and the dashboard is an admin's.
func TestTelemetryMiddleware(t *testing.T) {
	t.Parallel()
	telemetry, err := app.NewTelemetry(context.Background(), newSettings(t), "task-ui-test")
	assert.NoError(t, err)

	served := 0
	handler := telemetry.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		served++
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, enabled := range []bool{false, true} {
		opts := telemetry.Defaults()
		opts.Enabled = enabled
		assert.NoError(t, telemetry.Apply(opts))

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/runs", nil))
		assert.Equal(t, http.StatusNoContent, w.Code)
	}
	assert.Equal(t, 2, served)

	// A request with no user on it is refused, and oida answers a failed
	// Authorize with a 404 rather than announcing the dashboard.
	w := httptest.NewRecorder()
	telemetry.ServeHTTP(w, httptest.NewRequest(http.MethodGet, oida.DefaultPath, nil))
	assert.Equal(t, http.StatusNotFound, w.Code)

	// An admin reads it.
	r := httptest.NewRequest(http.MethodGet, oida.DefaultPath, nil)
	r = r.WithContext(access.WithUser(r.Context(), &model.User{Name: "black", Role: model.RoleAdmin}))
	w = httptest.NewRecorder()
	telemetry.ServeHTTP(w, r)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

// A nil Telemetry is an instance with no tracer, so a caller needs no
// check of its own.
func TestTelemetryNil(t *testing.T) {
	t.Parallel()
	var telemetry *app.Telemetry
	assert.Nil(t, telemetry.Tracer())

	served := 0
	handler := telemetry.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		served++
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, served)
}
