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
}
