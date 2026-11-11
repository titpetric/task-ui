package access_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestWithUser(t *testing.T) {
	t.Parallel()
	ctx := access.WithUser(context.Background(), &model.User{Name: "x"})
	assert.Equal(t, "x", access.UserFrom(ctx).Name)
	assert.Nil(t, access.UserFrom(context.Background()))
}
