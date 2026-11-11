//go:build integration

package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/storage"
)

// The storage finds a token by its hash, replaces a CLI one by name,
// touches it on use and drops it on delete.
func TestUserAuthTokenStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tokens := storage.NewUserAuthTokenStorage(newDB(t))

	assert.Error(t, tokens.Set(ctx, "u1", "", "h0"))
	assert.NoError(t, tokens.Set(ctx, "u1", "cli", "h1"))
	assert.NoError(t, tokens.Set(ctx, "u1", "other", "h2"))

	token, err := tokens.Find(ctx, "h1")
	assert.NoError(t, err)
	assert.Equal(t, "cli", token.Name)
	assert.Equal(t, model.TokenCLI, token.Kind)
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
	assert.Nil(t, list[0].ExpiresAt, "a token of Set never expires")

	assert.NoError(t, tokens.Delete(ctx, "u1", model.TokenCLI, "other"))
	_, err = tokens.Find(ctx, "h2")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = tokens.Get(ctx, "u1", model.TokenCLI, "other")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	assert.NoError(t, tokens.Set(ctx, "u2", "cli", "h6"))
	assert.NoError(t, tokens.DeleteByUser(ctx, "u1"))
	list, err = tokens.List(ctx, "u1")
	assert.NoError(t, err)
	assert.Len(t, list, 0)
	_, err = tokens.Find(ctx, "h6")
	assert.NoError(t, err, "another user's token stays")
}

// Add writes one more token of a kind and replaces nothing, so a user
// holds several MCP tokens at once; ListKind reads back their kind alone,
// newest first, and DeleteID drops one of them, scoped to its user.
func TestUserAuthTokenStorage_Many(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tokens := storage.NewUserAuthTokenStorage(newDB(t))
	// A stamp is kept to the millisecond, so an expiry compares equal
	// only when the test passes one already rounded.
	expires := time.Now().Add(time.Hour).UTC().Truncate(time.Millisecond)

	first, err := tokens.Add(ctx, "u1", model.TokenMCP, "mcp", "h1", &expires)
	assert.NoError(t, err)
	second, err := tokens.Add(ctx, "u1", model.TokenMCP, "mcp-2", "h2", &expires)
	assert.NoError(t, err)
	assert.NoError(t, tokens.Set(ctx, "u1", "cli", "h3"))

	// A name is unique per user and kind: the same one twice is refused,
	// and the CLI token takes a name an MCP token already holds.
	_, err = tokens.Add(ctx, "u1", model.TokenMCP, "mcp", "h4", &expires)
	assert.Error(t, err)
	_, err = tokens.Add(ctx, "u1", model.TokenCLI, "mcp", "h5", nil)
	assert.NoError(t, err)
	_, err = tokens.Add(ctx, "", model.TokenMCP, "mcp", "h6", &expires)
	assert.Error(t, err)

	list, err := tokens.ListKind(ctx, "u1", model.TokenMCP)
	assert.NoError(t, err)
	assert.Len(t, list, 2, "both MCP tokens stand, and no CLI one is among them")
	// Newest first, and two writes of one request fall on either side of a
	// millisecond as often as not, so which of the two leads is the stamps'
	// business. The later one never follows.
	ids := []string{list[0].ID, list[1].ID}
	assert.Contains(t, ids, first.ID)
	assert.Contains(t, ids, second.ID)
	assert.False(t, list[0].CreatedAt.Before(*list[1].CreatedAt))
	assert.True(t, expires.Equal(*list[0].ExpiresAt))

	got, err := tokens.Get(ctx, "u1", model.TokenMCP, "mcp-2")
	assert.NoError(t, err)
	assert.Equal(t, second.ID, got.ID)

	assert.NoError(t, tokens.DeleteID(ctx, "u2", second.ID), "another user's ID removes nothing")
	list, _ = tokens.ListKind(ctx, "u1", model.TokenMCP)
	assert.Len(t, list, 2)
	assert.NoError(t, tokens.DeleteID(ctx, "u1", second.ID))
	list, _ = tokens.ListKind(ctx, "u1", model.TokenMCP)
	assert.Len(t, list, 1)
	assert.Equal(t, first.ID, list[0].ID)
}

// DeleteKind removes the MCP tokens of a user and keeps those of
// `task-ui token`.
func TestUserAuthTokenStorage_DeleteKind(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tokens := storage.NewUserAuthTokenStorage(newDB(t))
	expires := time.Now().Add(time.Hour)
	_, err := tokens.Add(ctx, "u1", model.TokenMCP, "mcp", "h1", &expires)
	assert.NoError(t, err)
	assert.NoError(t, tokens.Set(ctx, "u1", "cli", "h2"))
	assert.NoError(t, tokens.DeleteKind(ctx, "u1", model.TokenMCP))
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
	add := func(userID, hash string, at time.Time) {
		t.Helper()
		_, err := tokens.Add(ctx, userID, model.TokenMCP, "mcp", hash, &at)
		assert.NoError(t, err)
	}
	assert.NoError(t, tokens.Set(ctx, "u1", "cli", "h1"))
	add("u1", "h2", now.Add(-time.Hour))
	add("u2", "h3", now.Add(-2*time.Hour))
	add("u3", "h4", now.Add(time.Hour))
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

// Every method returns the driver's error. A set deletes the token of the
// name before it writes the new one, so a dead pool stops it at the delete.
func TestUserAuthTokenStorage_Closed(t *testing.T) {
	t.Parallel()
	tokens := storage.NewUserAuthTokenStorage(closedDB(t))
	ctx := context.Background()
	expires := time.Now().Add(time.Hour)

	assert.ErrorContains(t, tokens.Set(ctx, "u1", "cli", "hash"), errClosed)
	_, err := tokens.Add(ctx, "u1", model.TokenMCP, "mcp", "hash", &expires)
	assert.ErrorContains(t, err, errClosed)
	_, err = tokens.Get(ctx, "u1", model.TokenCLI, "cli")
	assert.ErrorContains(t, err, errClosed)
	_, err = tokens.Find(ctx, "hash")
	assert.ErrorContains(t, err, errClosed)
	_, err = tokens.List(ctx, "u1")
	assert.ErrorContains(t, err, errClosed)
	_, err = tokens.ListKind(ctx, "u1", model.TokenMCP)
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, tokens.Touch(ctx, "t1"), errClosed)
	assert.ErrorContains(t, tokens.Delete(ctx, "u1", model.TokenCLI, "cli"), errClosed)
	assert.ErrorContains(t, tokens.DeleteID(ctx, "u1", "t1"), errClosed)
	assert.ErrorContains(t, tokens.DeleteKind(ctx, "u1", model.TokenMCP), errClosed)
	assert.ErrorContains(t, tokens.DeleteByUser(ctx, "u1"), errClosed)
}

// A token needs a user, a name and a hash.
func TestUserAuthTokenStorage_Refuses(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tokens := storage.NewUserAuthTokenStorage(newDB(t))

	assert.ErrorContains(t, tokens.Set(ctx, "", "cli", "hash"), "needs a user, a name and a hash")
	assert.ErrorContains(t, tokens.Set(ctx, "u1", "", "hash"), "needs a user, a name and a hash")
	assert.ErrorContains(t, tokens.Set(ctx, "u1", "cli", ""), "needs a user, a name and a hash")
}
