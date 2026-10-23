package permission

import (
	"strings"

	"github.com/titpetric/task-ui/config"
)

// Policy holds the entries of one subject, a role or a user: a path to
// Allow or Deny. A path it does not hold inherits.
type Policy map[string]string

// Effect returns the subject's entry for a path: its own, or for a rule
// of a section the section's all, a deny in either winning; "" when it
// sets neither. A rule a section adds, such as key/read, takes its own
// entry alone.
func (p Policy) Effect(path string) string {
	effect := p[path]
	if config.Default().IsExtra(path) {
		return effect
	}
	if section, _, ok := strings.Cut(path, "/"); ok {
		if all := p[section+"/"+config.All]; all == Deny || effect == "" {
			effect = all
		}
	}
	return effect
}
