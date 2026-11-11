package forgejo

import (
	"context"
	"io"
)

// Git clones and pulls the checkouts sync keeps, as git.Git does.
type Git interface {
	IsRepo(dir string) bool
	Clone(ctx context.Context, url, dir, branch string) (string, error)
	Pull(ctx context.Context, dir string, w io.Writer) error
}
