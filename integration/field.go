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
	// Env is the environment variable the setting sets in a run the
	// integration serves.
	Env string `json:"env,omitempty"`
	// Default is the value the setting has while its field is empty, and
	// what its form shows as the placeholder.
	Default string `json:"default,omitempty"`
	// Toggle gives the setting a switch: off leaves its variable out of a
	// run and keeps its value for the next switch on.
	Toggle bool `json:"toggle,omitempty"`
	// DefaultOn starts the switch on, for a variable a run needs.
	DefaultOn bool `json:"default_on,omitempty"`
	// Empty sends the variable with no value, such as ANTHROPIC_API_KEY,
	// and the form shows the switch alone.
	Empty bool `json:"empty,omitempty"`
	// Source takes the value from another field, and the form shows the
	// switch alone: a key that reaches a run as two variables.
	Source string `json:"source,omitempty"`
}

// Valid reports whether a value is one the setting takes: any without options, else one of them or empty.
func (f Field) Valid(value string) bool {
	return len(f.Options) == 0 || value == "" || slices.Contains(f.Options, value)
}
