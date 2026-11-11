package app

import (
	"context"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/frontend"
)

// Rules is `config/permissions.yml`, titles and path lists both.
func (h *FrontHost) Rules(context.Context) phpfront.Rules {
	file := config.Default()
	return phpfront.Rules{
		Rules:       frontRules(file.Rules),
		Sections:    frontRules(file.Sections),
		Permissions: frontRules(file.Permissions),
		Paths:       file.Paths(),
		Checks:      file.Checks(),
		Extras:      frontExtras(file),
	}
}

// frontRules is a titled list of the permission file.
func frontRules(entries []config.Entry) []phpfront.Rule {
	out := make([]phpfront.Rule, 0, len(entries))
	for _, entry := range entries {
		out = append(out, phpfront.Rule{Key: entry.Key, Title: entry.Title})
	}
	return out
}

// frontExtras is the rules a section added of its own, which a section's all
// does not stand for: keyed by the path a check names, titled as the rule
// titles itself. The matrix draws a row per one and composes the title with
// the section's, which it reads off Sections by the path's first segment.
func frontExtras(file *config.Permissions) []phpfront.Rule {
	out := make([]phpfront.Rule, 0, len(file.Extra))
	for _, extra := range file.Extra {
		out = append(out, phpfront.Rule{Key: extra.Path(), Title: extra.Title})
	}
	return out
}
