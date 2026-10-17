package access

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

type userKey struct{}

// UserFrom returns the user the guard put on the request context.
func UserFrom(ctx context.Context) *model.User {
	user, _ := ctx.Value(userKey{}).(*model.User)
	return user
}

// WithUser returns a context carrying a user, for handlers called
// outside the guard.
func WithUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}
