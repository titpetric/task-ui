package api

import (
	"context"
	"net/http"

	"github.com/titpetric/oida"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
)

// Telemetry is what the API writes: the options in use, and a save. It
// is app.Telemetry, as an interface, so this package stays off the
// composition root.
type Telemetry interface {
	Load(ctx context.Context) (oida.Options, error)
	Save(ctx context.Context, opts oida.Options) error
}

// TelemetryRequest is the Telemetry settings page's form.
//
// Every field is a pointer, so a request that names one setting changes
// one setting and the rest keep what they have: the page posts the whole
// form, and a script or an agent posts the switch alone.
//
// The names are the yaml names of oida.Options, which is what the stored
// row holds, so the form, the setting and the library all spell a field
// the same way.
type TelemetryRequest struct {
	Enabled          *bool    `json:"enabled"`
	ServiceName      *string  `json:"service_name"`
	SampleRate       *float64 `json:"sample_rate"`
	RingBufferSize   *int     `json:"ring_buffer_size"`
	TopRequests      *int     `json:"top_requests"`
	MaxSpansPerTrace *int     `json:"max_spans_per_trace"`
	RefreshInterval  *int     `json:"refresh_interval"`
	LiveStream       *bool    `json:"live_stream"`
	CaptureLogs      *bool    `json:"capture_logs"`
	TrackMemoryUse   *bool    `json:"track_memory_use"`
	TrustRequestID   *bool    `json:"trust_request_id"`
	Path             *string  `json:"path"`
	IgnorePaths      *string  `json:"ignore_paths"`
	AllowedNetworks  *string  `json:"allowed_networks"`
}

// TelemetryView is the settings as the API answers them.
//
// oida.Options is not written out directly: every tag on it is a yaml
// tag and five of its fields are plain funcs, so encoding/json refuses
// the struct with "unsupported type: func(*http.Request) string". The
// names here are its yaml names, so the form, the stored row and the
// answer all spell a setting the same way.
type TelemetryView struct {
	Enabled          bool     `json:"enabled"`
	ServiceName      string   `json:"service_name"`
	SampleRate       float64  `json:"sample_rate"`
	RingBufferSize   int      `json:"ring_buffer_size"`
	TopRequests      int      `json:"top_requests"`
	MaxSpansPerTrace int      `json:"max_spans_per_trace"`
	RefreshInterval  int      `json:"refresh_interval"`
	LiveStream       bool     `json:"live_stream"`
	CaptureLogs      bool     `json:"capture_logs"`
	TrackMemoryUse   bool     `json:"track_memory_use"`
	TrustRequestID   bool     `json:"trust_request_id"`
	Path             string   `json:"path"`
	IgnorePaths      []string `json:"ignore_paths"`
	AllowedNetworks  []string `json:"allowed_networks"`
}

// NewTelemetryView returns the options as the API answers them.
func NewTelemetryView(opts oida.Options) TelemetryView {
	return TelemetryView{
		Enabled: opts.Enabled, ServiceName: opts.ServiceName, SampleRate: opts.SampleRate,
		RingBufferSize: opts.RingBufferSize, TopRequests: opts.TopRequests,
		MaxSpansPerTrace: opts.MaxSpansPerTrace, RefreshInterval: opts.RefreshInterval,
		LiveStream: opts.LiveStream, CaptureLogs: opts.CaptureLogs,
		TrackMemoryUse: opts.TrackMemoryUse, TrustRequestID: opts.TrustRequestID,
		Path: opts.Path, IgnorePaths: opts.IgnorePaths, AllowedNetworks: opts.AllowedNetworks,
	}
}

// saveTelemetry sets what the instance records of its own work, for
// admins alone. Enabled on its own keeps the traces already recorded;
// any other change rebuilds the recorder and empties them. It answers the
// options as they stand.
func (h *Handlers) saveTelemetry(w http.ResponseWriter, r *http.Request) {
	if !access.UserFrom(r.Context()).CanAdmin() {
		fail(w, http.StatusForbidden, "only an admin changes the telemetry settings")
		return
	}
	if h.opts.Telemetry == nil {
		fail(w, http.StatusNotFound, "telemetry isn't set up on this instance")
		return
	}
	var req TelemetryRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	opts, err := h.opts.Telemetry.Load(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	req.Apply(&opts)
	// A sample rate or a path out of range is the caller's mistake, and
	// the library says which. Validating here rather than reading it off
	// the save keeps a storage failure a 500 and this a 400.
	if err := opts.Validate(); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.opts.Telemetry.Save(r.Context(), opts); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, NewTelemetryView(opts))
}

// Apply writes the fields the request named onto the options it has, and
// leaves every other field as it stands.
func (t TelemetryRequest) Apply(opts *oida.Options) {
	setIf(&opts.Enabled, t.Enabled)
	setIf(&opts.ServiceName, t.ServiceName)
	setIf(&opts.SampleRate, t.SampleRate)
	setIf(&opts.RingBufferSize, t.RingBufferSize)
	setIf(&opts.TopRequests, t.TopRequests)
	setIf(&opts.MaxSpansPerTrace, t.MaxSpansPerTrace)
	setIf(&opts.RefreshInterval, t.RefreshInterval)
	setIf(&opts.LiveStream, t.LiveStream)
	setIf(&opts.CaptureLogs, t.CaptureLogs)
	setIf(&opts.TrackMemoryUse, t.TrackMemoryUse)
	setIf(&opts.TrustRequestID, t.TrustRequestID)
	setIf(&opts.Path, t.Path)
	// The two list fields arrive as the textareas that edit them, one
	// entry a line, and a blank line is not an entry: an empty string in
	// IgnorePaths would match every path.
	if t.IgnorePaths != nil {
		opts.IgnorePaths = model.Lines(*t.IgnorePaths)
	}
	if t.AllowedNetworks != nil {
		opts.AllowedNetworks = model.Lines(*t.AllowedNetworks)
	}
}

// setIf writes a value the request named, and leaves the field alone
// where it named none.
func setIf[T any](field *T, value *T) {
	if value != nil {
		*field = *value
	}
}

// listTelemetry answers the options in use, for a user who may see a step
// of the settings.
func (h *Handlers) listTelemetry(w http.ResponseWriter, r *http.Request) {
	if h.opts.Telemetry == nil {
		fail(w, http.StatusNotFound, "telemetry isn't set up on this instance")
		return
	}
	opts, err := h.opts.Telemetry.Load(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, NewTelemetryView(opts))
}
