package permission

import (
	"strings"

	"github.com/titpetric/task-ui/config"
)

// Policy holds the entries of one subject, a role or a user: a path to
// Allow or Deny. A path it does not hold inherits.
type Policy map[string]string

// Effect returns the subject's entry for a path: its own, or for a rule
// of a section the section's all; "" when it sets neither.
func (p Policy) Effect(path string) string {
	if effect, ok := p[path]; ok {
		return effect
	}
	if section, _, ok := strings.Cut(path, "/"); ok {
		return p[section+"/"+config.All]
	}
	return ""
}
