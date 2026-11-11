package phpfront_test

import (
	"context"

	"github.com/titpetric/task-ui/frontend"
)

// What the fixture answers of the settings: the keys, the instance, the
// retention, the integrations and the vault, grouped the way
// host_settings.go declares the shapes.

func (h *host) Keys(context.Context) phpfront.Keystore {
	return phpfront.Keystore{}
}

func (h *host) KeyFolders(context.Context) []phpfront.KeyFolder {
	return nil
}

func (h *host) Instance(context.Context, bool) phpfront.Instance {
	return phpfront.Instance{}
}

func (h *host) Usage(context.Context) phpfront.Retention {
	return phpfront.Retention{}
}

func (h *host) Telemetry(context.Context) phpfront.Telemetry {
	return phpfront.Telemetry{}
}

// A service answers its registry entry and what its settings hold; the token
// it is set up with is not in the answer at all.
func (h *host) Services(context.Context) []phpfront.Integration {
	return []phpfront.Integration{{
		ID: "claude", Title: "Claude Code", Description: "The token Claude Code signs in with.",
		Credential: "OAuth token", Steps: []string{"Run claude setup-token."},
		Fields: []phpfront.IntegrationField{
			{Name: "token", Label: "OAuth token", Secret: true, Env: "CLAUDE_CODE_OAUTH_TOKEN"},
			{Name: "who", Label: "Who", Optional: true, Who: true},
		},
		Configured: true,
		Values:     map[string]string{"who": "admin, ops"},
		Set:        map[string]bool{"token": true},
		Checked:    []string{"admin", "ops"},
		Enabled:    map[string]bool{},
	}}
}

// The vault answers the variables it opened and the first one it would not,
// which is the shape the page draws: the table and the error above it.
func (h *host) Variables(context.Context) phpfront.Variables {
	return phpfront.Variables{Values: map[string]string{"TOKEN": "", "REGION": "eu"}}
}

// One variable as its form opens it. A secret draws no value, so the row the
// form edits answers content only where the variable holds some.
func (h *host) Variable(_ context.Context, id string) phpfront.Variable {
	if id != "REGION" {
		return phpfront.Variable{}
	}
	return phpfront.Variable{Content: "eu"}
}

// The environment the integrations add is the reader's: a plain variable and a
// credential, which answers its name and no value.
func (h *host) IntegrationEnv(context.Context) []phpfront.EnvVar {
	return []phpfront.EnvVar{
		{Name: "CLAUDE_CODE_OAUTH_TOKEN", Redacted: true},
		{Name: "CLAUDE_MODEL", Value: "opus"},
	}
}

func (h *host) Modules(context.Context) []phpfront.Plugin {
	return nil
}

func (h *host) Entries(context.Context, []string) []phpfront.Project {
	return nil
}
