package view

import (
	"strings"

	"github.com/titpetric/task-ui/integration"
)

// EnvVar is one variable an integration sets in the environment of a
// run, its value redacted when its name holds a credential.
type EnvVar struct {
	Name     string
	Value    string
	Redacted bool
}

// Integration is an integration as its form shows it: the plain settings
// as they are, the secret ones only as set or not.
type Integration struct {
	integration.Kind
	Configured bool
	Values     map[string]string
	Set        map[string]bool
	// Checked are the roles and users a Who setting names.
	Checked map[string]bool
	// Enabled are the switches of the settings that have one.
	Enabled map[string]bool
}

// NewIntegration reads an integration's settings for its form.
func NewIntegration(kind integration.Kind, values map[string]string) Integration {
	view := Integration{Kind: kind, Configured: integration.Configured(kind, values), Values: map[string]string{}, Set: map[string]bool{}, Checked: map[string]bool{}, Enabled: map[string]bool{}}
	for _, field := range kind.Fields {
		if field.Toggle {
			view.Enabled[field.Name] = integration.Enabled(field, values)
		}
		if field.Secret {
			view.Set[field.Name] = values[field.Name] != ""
			continue
		}
		view.Values[field.Name] = values[field.Name]
		if field.Who {
			for _, name := range strings.Split(values[field.Name], ",") {
				if name = strings.TrimSpace(name); name != "" {
					view.Checked[name] = true
				}
			}
		}
	}
	return view
}
