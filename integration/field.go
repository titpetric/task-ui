package integration

import (
	"slices"
)

// Field is one setting of an integration, as its form shows it.
type Field struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Help        string `json:"help,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	// Secret keeps the value out of every page and response: a form shows
	// whether it has a value, and an empty field on save keeps it.
	Secret bool `json:"secret,omitempty"`
	// Optional leaves the integration set up without a value.
	Optional bool `json:"optional,omitempty"`
	// Who makes the value a comma separated list of roles and users, which
	// the form shows as a box each.
	Who bool `json:"who,omitempty"`
	// Options are the values the setting takes, the first the default,
	// which the form shows as a list to pick from.
	Options []string `json:"options,omitempty"`
	// Generated fills an empty secret with a random value the first time
	// the integration saves, and keeps it after.
	Generated bool `json:"generated,omitempty"`
}

// Valid reports whether a value is one the setting takes: any without options, else one of them or empty.
func (f Field) Valid(value string) bool {
	return len(f.Options) == 0 || value == "" || slices.Contains(f.Options, value)
}
