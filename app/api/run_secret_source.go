package api

import (
	"context"
)

// RunSecretSource finds a secret for the run that holds a token, as
// runner.Manager does.
type RunSecretSource interface {
	Secret(ctx context.Context, token, name string) (string, error)
}
