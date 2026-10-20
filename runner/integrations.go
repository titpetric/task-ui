package runner

import (
	"context"
)

// Integrations return the variables the integrations add to a run started
// by a user of a role.
type Integrations interface {
	Env(ctx context.Context, user, role string) (map[string]string, error)
}
