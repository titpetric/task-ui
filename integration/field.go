package integration

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
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
	// WhoEmpty words what nothing checked means for a Who field; empty
	// keeps "every run gets the token".
	WhoEmpty string `json:"who_empty,omitempty"`
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
	// Path makes the instance own the field: its value is the instance
	// URL and this path, a form shows it and takes no input for it, and
	// a value sent for it is dropped.
	Path string `json:"path,omitempty"`
}

// Fixed reports whether the instance owns the field's value, so a form shows it and saves nothing for it.
func (f Field) Fixed() bool {
	return f.Path != ""
}

// Valid reports whether a value is one the setting takes: any without options, else one of them or empty.
func (f Field) Valid(value string) bool {
	return len(f.Options) == 0 || value == "" || slices.Contains(f.Options, value)
}

// Check returns why the field refuses a value, or nil: one its options do not name, or an address a field with a Path cannot use.
func (f Field) Check(value string) error {
	if !f.Valid(value) {
		return fmt.Errorf("%q for %s: one of %s", value, f.Name, strings.Join(f.Options, ", "))
	}
	if f.Path == "" || value == "" {
		return nil
	}
	address, err := url.Parse(value)
	if err != nil || address.Host == "" || (address.Scheme != "http" && address.Scheme != "https") {
		return fmt.Errorf("%q for %s: an http or https address, or empty to take the instance URL", value, f.Name)
	}
	return nil
}

// Value returns what a form shows in the field: the stored value, else the instance URL and Path, else Default.
func (f Field) Value(stored, baseURL string) string {
	if stored != "" {
		return stored
	}
	if f.Path != "" && baseURL != "" {
		return strings.TrimRight(baseURL, "/") + f.Path
	}
	return f.Default
}
