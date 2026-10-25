//go:build integration

package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// The storage finds a session by its hash, touches it on use, and drops
// it by hash, by user, or once it expired.
func TestUserSessionStorage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sessions := storage.NewUserSessionStorage(newDB(t))
	now := time.Now()

	_, err := sessions.Create(ctx, "u1", "", now.Add(time.Hour))
	assert.Error(t, err)
	created, err := sessions.Create(ctx, "u1", "h1", now.Add(time.Hour))
	assert.NoError(t, err)
	_, err = sessions.Create(ctx, "u1", "h1", now.Add(time.Hour))
	assert.Error(t, err, "duplicate hash")
	_, err = sessions.Create(ctx, "u1", "h2", now.Add(time.Hour))
	assert.NoError(t, err)
	_, err = sessions.Create(ctx, "u2", "h3", now.Add(time.Hour))
	assert.NoError(t, err)
	_, err = sessions.Create(ctx, "u2", "old", now.Add(-time.Minute))
	assert.NoError(t, err)

	session, err := sessions.Find(ctx, "h1")
	assert.NoError(t, err)
	assert.Equal(t, created.ID, session.ID)
	assert.Equal(t, "u1", session.UserID)
	assert.Nil(t, session.UsedAt)
	assert.NoError(t, sessions.Touch(ctx, session.ID))
	session, _ = sessions.Find(ctx, "h1")
	assert.NotNil(t, session.UsedAt)

	assert.NoError(t, sessions.DeleteExpired(ctx, now))
	_, err = sessions.Find(ctx, "old")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = sessions.Find(ctx, "h3")
	assert.NoError(t, err, "a live session stays")

	assert.NoError(t, sessions.Delete(ctx, "h3"))
	_, err = sessions.Find(ctx, "h3")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	assert.NoError(t, sessions.DeleteByUser(ctx, "u1"))
	_, err = sessions.Find(ctx, "h1")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = sessions.Find(ctx, "h2")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

// Expired lists the sessions past their expiry, which DeleteExpired removes.
func TestUserSessionStorage_Expired(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sessions := storage.NewUserSessionStorage(newDB(t))
	now := time.Now()
	_, err := sessions.Create(ctx, "u1", "live", now.Add(time.Hour))
	assert.NoError(t, err)
	old, err := sessions.Create(ctx, "u1", "old", now.Add(-time.Minute))
	assert.NoError(t, err)
	expired, err := sessions.Expired(ctx, now)
	assert.NoError(t, err)
	assert.Len(t, expired, 1)
	assert.Equal(t, old.ID, expired[0].ID)
	assert.NoError(t, sessions.DeleteExpired(ctx, now))
	expired, _ = sessions.Expired(ctx, now)
	assert.Len(t, expired, 0)
	_, err = sessions.Find(ctx, "live")
	assert.NoError(t, err)
}
