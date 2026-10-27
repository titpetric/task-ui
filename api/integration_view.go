package api

import (
	"github.com/titpetric/task-ui/integration"
)

// IntegrationView is an integration as Settings shows it: its settings,
// the secret ones only as set or not.
type IntegrationView struct {
	integration.Kind
	Configured bool              `json:"configured"`
	Values     map[string]string `json:"values"`
	// Set names the secret settings that have a value.
	Set []string `json:"set"`
	// Switches are the settings with a switch, by name, on or off.
	Switches map[string]bool `json:"switches"`
}
