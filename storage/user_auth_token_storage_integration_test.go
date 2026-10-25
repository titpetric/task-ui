//go:build integration

package storage_test

import (
	"context"
	"testing"
	"time"

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
	_, err = tokens.Get(ctx, "u1", "mcp")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.Nil(t, list[0].ExpiresAt, "a token of Set never expires")

	// SetUntil keeps an expiry; Get finds the token by name.
	expires := time.Date(2026, time.November, 21, 10, 0, 0, 0, time.UTC)
	row, err := tokens.SetUntil(ctx, "u1", "mcp", "h4", expires)
	assert.NoError(t, err)
	assert.Equal(t, expires, *row.ExpiresAt)
	got, err := tokens.Get(ctx, "u1", "mcp")
	assert.NoError(t, err)
	assert.Equal(t, row.ID, got.ID)
	assert.True(t, expires.Equal(*got.ExpiresAt))
	_, err = tokens.SetUntil(ctx, "", "mcp", "h5", expires)
	assert.Error(t, err)

	assert.NoError(t, tokens.Set(ctx, "u2", "cli", "h6"))
	assert.NoError(t, tokens.DeleteByUser(ctx, "u1"))
	list, err = tokens.List(ctx, "u1")
	assert.NoError(t, err)
	assert.Len(t, list, 0)
	_, err = tokens.Find(ctx, "h6")
	assert.NoError(t, err, "another user's token stays")
}

// DeleteMCP removes the tokens that expire, the MCP ones, and keeps
// those of `task-ui token`.
func TestUserAuthTokenStorage_DeleteMCP(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tokens := storage.NewUserAuthTokenStorage(newDB(t))
	_, err := tokens.SetUntil(ctx, "u1", "mcp", "h1", time.Now().Add(time.Hour))
	assert.NoError(t, err)
	assert.NoError(t, tokens.Set(ctx, "u1", "cli", "h2"))
	assert.NoError(t, tokens.DeleteMCP(ctx, "u1"))
	list, err := tokens.List(ctx, "u1")
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "cli", list[0].Name)
}

// Expired lists the tokens past their expiry, oldest expiry first;
// DeleteExpired removes them and leaves the live ones and the tokens
// without an expiry.
func TestUserAuthTokenStorage_Expired(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tokens := storage.NewUserAuthTokenStorage(newDB(t))
	now := time.Now()
	assert.NoError(t, tokens.Set(ctx, "u1", "cli", "h1"))
	_, err := tokens.SetUntil(ctx, "u1", "mcp", "h2", now.Add(-time.Hour))
	assert.NoError(t, err)
	_, err = tokens.SetUntil(ctx, "u2", "mcp", "h3", now.Add(-2*time.Hour))
	assert.NoError(t, err)
	_, err = tokens.SetUntil(ctx, "u3", "mcp", "h4", now.Add(time.Hour))
	assert.NoError(t, err)
	expired, err := tokens.Expired(ctx, now)
	assert.NoError(t, err)
	assert.Len(t, expired, 2)
	assert.Equal(t, "h3", expired[0].Hash)
	assert.NoError(t, tokens.DeleteExpired(ctx, now))
	expired, _ = tokens.Expired(ctx, now)
	assert.Len(t, expired, 0)
	for _, hash := range []string{"h1", "h4"} {
		_, err := tokens.Find(ctx, hash)
		assert.NoError(t, err, hash)
	}
}
