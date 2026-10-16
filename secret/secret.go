// Package secret keeps the variables task-ui adds to runs: for every
// folder, or for one.
package secret

import (
	"errors"
	"regexp"

	"github.com/titpetric/task-ui/folder"
)

// ScopeAll is the scope that applies to every folder.
const ScopeAll = "all"

// Secret is one environment variable task-ui adds to runs: to every
// folder when Scope is ScopeAll, or to the folder Scope names.
type Secret struct {
	Scope string `json:"scope"`
	Key   string `json:"key"`
	Value string `json:"-"`
}

var validKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ErrInvalidKey is returned for a variable name a shell would not accept.
var ErrInvalidKey = errors.New("a variable name is letters, digits and underscores, and does not start with a digit")

// ValidKey reports whether key is a variable name.
func ValidKey(key string) bool {
	return validKey.MatchString(key)
}

// ValidScope reports whether scope is ScopeAll or a folder ID.
func ValidScope(scope string) bool {
	return scope == ScopeAll || folder.ValidID(scope)
}
