package runner

import (
	"context"
)

// Integrations return the variables the integrations add to a run of a
// folder.
type Integrations interface {
	Env(ctx context.Context, folder string) (map[string]string, error)
}
