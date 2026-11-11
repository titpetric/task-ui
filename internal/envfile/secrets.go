package envfile

import (
	"context"
)

// Secrets returns the secrets a folder's runs get, which a written env
// file carries after its content.
type Secrets interface {
	For(ctx context.Context, folder string) (map[string]string, error)
}
