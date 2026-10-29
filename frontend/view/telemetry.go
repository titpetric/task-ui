package view

import (
	"strings"

	"github.com/titpetric/oida"
)

// Telemetry is the settings page of the tracer: the options in use, and
// the two list fields as the textareas that edit them spell them.
type Telemetry struct {
	Options oida.Options

	// Path is where the dashboard is mounted, which the router read at
	// startup. A path saved after that takes a restart, and the page says
	// so, so this is what the link on the page points at rather than
	// whatever the form holds.
	Path string

	// Admin says the user may save. Anyone who may open a step of the
	// settings reads the page.
	Admin bool
}

// IgnorePaths is the paths never traced, one per line, for the textarea.
func (t *Telemetry) IgnorePaths() string {
	return strings.Join(t.Options.IgnorePaths, "\n")
}

// AllowedNetworks is the CIDR ranges that may open the dashboard, one per
// line. Empty is every network, which the page's help says.
func (t *Telemetry) AllowedNetworks() string {
	return strings.Join(t.Options.AllowedNetworks, "\n")
}
