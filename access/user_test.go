package access_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestWithUser(t *testing.T) {
	t.Parallel()
	ctx := access.WithUser(context.Background(), &model.User{Name: "x"})
	assert.Equal(t, "x", access.UserFrom(ctx).Name)
}
