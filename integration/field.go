package integration

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
}
