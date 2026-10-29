package app

import (
	"context"
	"net/http"
	"sync/atomic"

	chi "github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"

	"github.com/titpetric/oida"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/storage"
)

// TelemetrySetting is the name the options live under in the setting
// table, beside the retention rules and the parent environment.
const TelemetrySetting = "telemetry.options"

// Telemetry is the instance's tracer and the settings page behind it.
//
// The tracer sits behind an atomic pointer because oida builds one from
// its options and exposes no way to change them after: Apply builds a new
// tracer and swaps it, so a save takes effect without a restart. The
// middleware and the dashboard both read through the pointer, so neither
// holds a tracer the settings have replaced.
//
// Enabled is the exception. A tracer can be switched on and off in place,
// so flipping that alone keeps every trace recorded so far; any other
// change starts a fresh ring buffer, which the page says.
type Telemetry struct {
	current  atomic.Pointer[oida.Tracer]
	settings *storage.SettingStorage
	service  string
}

// NewTelemetry returns the telemetry of an instance, with a tracer built
// from the options stored under TelemetrySetting. An instance with none
// stored takes oida's defaults, which read the OIDA_* environment, so a
// first start is configured from the deployment and the page after that.
func NewTelemetry(ctx context.Context, settings *storage.SettingStorage, service string) (*Telemetry, error) {
	t := &Telemetry{settings: settings, service: service}
	opts, err := t.Load(ctx)
	if err != nil {
		return nil, err
	}
	if err := t.build(opts); err != nil {
		return nil, err
	}
	return t, nil
}

// Defaults are the options of an instance that stored none: oida's own,
// named for this service, with the debug front end behind an admin and
// the statistics grouped by chi's routed pattern.
func (t *Telemetry) Defaults() oida.Options {
	opts := oida.NewOptions(t.service)
	opts.RouteFunc = routePattern
	opts.Authorize = authorizeTelemetry
	return opts
}

// Load returns the stored options, over the defaults. A value that does
// not parse is reported rather than served, so a hand-edited row cannot
// leave the instance with a tracer nobody configured.
func (t *Telemetry) Load(ctx context.Context) (oida.Options, error) {
	opts := t.Defaults()
	if t.settings == nil {
		return opts, nil
	}
	values, err := t.settings.Values(ctx)
	if err != nil {
		return opts, err
	}
	stored, ok := values[TelemetrySetting]
	if !ok || stored == "" {
		return opts, nil
	}
	// The stored row is the yaml-tagged fields of oida.Options, which are
	// the configurable ones: the sampler, the storage driver and the five
	// callbacks carry `yaml:"-"`, so a document cannot name them and
	// unmarshalling over the defaults leaves them as they are.
	//
	// YAML and not JSON. Every tag on oida.Options is a yaml tag, and the
	// callbacks are plain func fields, so encoding/json refuses the struct
	// outright with "unsupported type: func(*http.Request) string".
	if err := yaml.Unmarshal([]byte(stored), &opts); err != nil {
		return t.Defaults(), err
	}
	// ReadEnv applies the OIDA_* environment when the tracer is built,
	// which would override what the page just saved. A stored row is the
	// instance's own answer, so the environment stops deciding.
	opts.ReadEnv = false
	opts.RouteFunc = routePattern
	opts.Authorize = authorizeTelemetry
	return opts, nil
}

// Save stores the options and applies them. Enabled alone goes through
// the running tracer and keeps the traces it holds; anything else builds
// a tracer and swaps it, which starts an empty ring buffer.
func (t *Telemetry) Save(ctx context.Context, opts oida.Options) error {
	if err := opts.Validate(); err != nil {
		return err
	}
	data, err := yaml.Marshal(opts)
	if err != nil {
		return err
	}
	if t.settings != nil {
		if err := t.settings.Set(ctx, map[string]string{TelemetrySetting: string(data)}); err != nil {
			return err
		}
	}
	return t.Apply(opts)
}

// Apply swaps the tracer for one built from opts, or switches the one
// running on and off when that is the only change.
func (t *Telemetry) Apply(opts oida.Options) error {
	if running := t.Tracer(); running != nil && sameButEnabled(running.Options(), opts) {
		running.SetEnabled(opts.Enabled)
		return nil
	}
	return t.build(opts)
}

// Tracer is the tracer in use, which the middleware and the dashboard
// read through so neither holds one the settings replaced.
func (t *Telemetry) Tracer() *oida.Tracer {
	if t == nil {
		return nil
	}
	return t.current.Load()
}

// Path is where the dashboard is mounted, which a save cannot change: the
// router reads it once, at startup, and a path saved after that takes a
// restart. The settings page says so.
func (t *Telemetry) Path() string {
	if tracer := t.Tracer(); tracer != nil {
		if path := tracer.Options().Path; path != "" {
			return path
		}
	}
	return oida.DefaultPath
}

// Middleware records every sampled request. It reads the pointer per
// request, so a save applies to the next one.
func (t *Telemetry) Middleware(next http.Handler) http.Handler {
	if t == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := t.Tracer()
		if tracer == nil {
			next.ServeHTTP(w, r)
			return
		}
		tracer.Middleware(next).ServeHTTP(w, r)
	})
}

// ServeHTTP serves the debug front end, reading the pointer per request
// for the same reason the middleware does.
func (t *Telemetry) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tracer := t.Tracer()
	if tracer == nil {
		http.NotFound(w, r)
		return
	}
	tracer.ServeHTTP(w, r)
}

// build makes a tracer from opts and puts it in place of the one before.
func (t *Telemetry) build(opts oida.Options) error {
	tracer, err := oida.New(opts)
	if err != nil {
		return err
	}
	t.current.Store(tracer)
	return nil
}

// sameButEnabled reports whether two option sets differ in Enabled and
// nothing else, which is the one change a running tracer can take.
//
// It compares the yaml-tagged fields, because those are the ones a save
// can change: the callbacks are the instance's own and the same function
// either way.
func sameButEnabled(running, next oida.Options) bool {
	running.Enabled = next.Enabled
	before, err := yaml.Marshal(running)
	if err != nil {
		return false
	}
	after, err := yaml.Marshal(next)
	if err != nil {
		return false
	}
	return string(before) == string(after)
}

// routePattern groups the statistics by the routed pattern rather than
// the request path, so /runs/1 and /runs/2 aggregate into GET /runs/{id}.
//
// A catch-all mount would group every request it serves under one pattern
// that says nothing, so those patterns are dropped and the requests group
// by path. Without that, every path of every module would read as one row
// under /m/{vendor}/{plugin}/*.
func routePattern(r *http.Request) string {
	routeContext := chi.RouteContext(r.Context())
	if routeContext == nil {
		return ""
	}
	switch pattern := routeContext.RoutePattern(); pattern {
	case "/*", "/", "":
		return ""
	default:
		return pattern
	}
}

// authorizeTelemetry gates the debug front end on an admin of the
// instance. A trace holds request paths, query strings and the errors a
// handler recorded, which is more than a viewer may read.
//
// A request with no user on it is the sign in page's own, and refused:
// oida answers a failed Authorize with a 404, so the dashboard does not
// announce itself to whoever asks.
func authorizeTelemetry(r *http.Request) bool {
	user := access.UserFrom(r.Context())
	return user != nil && user.CanAdmin()
}
