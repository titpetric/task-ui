package frontend

import (
	"context"
	"net/http"

	"github.com/titpetric/oida"

	"github.com/titpetric/task-ui/frontend/view"
)

// Telemetry is what the settings page reads: the options the tracer runs
// on, and where the dashboard is mounted. app.Telemetry implements it,
// which keeps this package off the composition root.
type Telemetry interface {
	// Load returns the options in use, the stored ones over the defaults.
	Load(ctx context.Context) (oida.Options, error)

	// Path is where the dashboard is mounted. The router read it at
	// startup, so a path saved after that takes a restart.
	Path() string
}

// telemetry shows the one form that configures what the instance records
// of its own work: the switch, the sample rate, how much it retains, and
// who may read the traces.
//
// Anyone who may open a step of the settings reads it; an admin saves it.
func (h *Handlers) telemetry(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, view.StepTelemetry)
	if !p.CanSettings() {
		h.refuse(w, r)
		return
	}
	p.Crumbs = view.Crumbs("Settings", "/settings", view.StepTelemetry)
	p.Step = view.StepTelemetry

	// An instance with no telemetry wired shows the defaults, read only,
	// rather than an empty page: the fields are what it would record.
	shown := &view.Telemetry{Options: oida.NewOptions("task-ui"), Path: oida.DefaultPath}
	if h.opts.Telemetry != nil {
		opts, err := h.opts.Telemetry.Load(r.Context())
		if err != nil {
			http.Error(w, "reading the telemetry settings failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		shown.Options, shown.Path = opts, h.opts.Telemetry.Path()
	}
	shown.Admin = p.User != nil && p.User.CanAdmin() && h.opts.Telemetry != nil
	p.Telemetry = shown
	h.render(w, http.StatusOK, "telemetry", p)
}
