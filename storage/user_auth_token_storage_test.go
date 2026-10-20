package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// The storage finds a token by its hash, replaces it by name, touches it
// on use and drops it on delete.
func TestUserAuthTokenStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tokens := storage.NewUserAuthTokenStorage(newDB(t))

	assert.Error(t, tokens.Set(ctx, "u1", "", "h0"))
	assert.NoError(t, tokens.Set(ctx, "u1", "cli", "h1"))
	assert.NoError(t, tokens.Set(ctx, "u1", "mcp", "h2"))

	token, err := tokens.Find(ctx, "h1")
	assert.NoError(t, err)
	assert.Equal(t, "cli", token.Name)
	assert.Nil(t, token.UsedAt)
	assert.NoError(t, tokens.Touch(ctx, token.ID))
	token, _ = tokens.Find(ctx, "h1")
	assert.NotNil(t, token.UsedAt)

	// Set under the same name replaces: the old hash fails to sign in.
	assert.NoError(t, tokens.Set(ctx, "u1", "cli", "h3"))
	_, err = tokens.Find(ctx, "h1")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	list, err := tokens.List(ctx, "u1")
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "cli", list[0].Name)

	assert.NoError(t, tokens.Delete(ctx, "u1", "mcp"))
	_, err = tokens.Find(ctx, "h2")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}
