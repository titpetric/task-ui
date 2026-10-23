//go:build integration

package access_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestNewSessions(t *testing.T) {
	t.Parallel()
	guard, _ := newGuard(t)
	assert.NotNil(t, guard.Sessions.Store)
	assert.NotNil(t, guard.Sessions.Users)
	assert.Equal(t, time.Hour, guard.Sessions.TTL())
}

// A token is opaque: random base64url, not a JWT, and the database holds
// its hash alone.
func TestSessions_Issue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	guard, ops := newGuard(t)
	s := guard.Sessions

	token, err := s.Issue(ctx, ops)
	assert.NoError(t, err)
	assert.Len(t, token, 43)
	assert.False(t, strings.Contains(token, "."), "no JWT segments")
	other, err := s.Issue(ctx, ops)
	assert.NoError(t, err)
	assert.NotEqual(t, token, other)

	_, err = s.Store.Find(ctx, token)
	assert.ErrorIs(t, err, storage.ErrNotFound, "the token itself is not stored")
	row, err := s.Store.Find(ctx, access.HashToken(token))
	assert.NoError(t, err)
	assert.Equal(t, ops.ID, row.UserID)
	assert.NotNil(t, row.ExpiresAt)
	assert.True(t, row.ExpiresAt.After(time.Now().Add(59*time.Minute)))

	// Issue prunes the sessions that expired.
	s.Now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	_, err = s.Issue(ctx, ops)
	assert.NoError(t, err)
	_, err = s.Store.Find(ctx, access.HashToken(token))
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestSessions_Resolve(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	guard, ops := newGuard(t)
	s := guard.Sessions
	token, err := s.Issue(ctx, ops)
	assert.NoError(t, err)

	user, err := s.Resolve(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, ops.ID, user.ID)
	row, _ := s.Store.Find(ctx, access.HashToken(token))
	assert.NotNil(t, row.UsedAt)
	used := *row.UsedAt

	// A second look within the minute writes nothing.
	_, err = s.Resolve(ctx, token)
	assert.NoError(t, err)
	row, _ = s.Store.Find(ctx, access.HashToken(token))
	assert.Equal(t, used, *row.UpdatedAt)

	for _, bad := range []string{"", "nope", token + "x", strings.ToUpper(token), access.HashToken(token)} {
		_, err = s.Resolve(ctx, bad)
		assert.ErrorIs(t, err, access.ErrInvalidToken, bad)
	}

	// A session of a user missing from storage signs in nobody.
	stray, err := s.Issue(ctx, &model.User{ID: model.NewID(), Name: "gone"})
	assert.NoError(t, err)
	_, err = s.Resolve(ctx, stray)
	assert.ErrorIs(t, err, access.ErrInvalidToken)

	s.Now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	_, err = s.Resolve(ctx, token)
	assert.ErrorIs(t, err, access.ErrInvalidToken, "expired")
}

func TestSessions_Revoke(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	guard, ops := newGuard(t)
	s := guard.Sessions
	token, err := s.Issue(ctx, ops)
	assert.NoError(t, err)

	assert.NoError(t, s.Revoke(ctx, token))
	_, err = s.Resolve(ctx, token)
	assert.ErrorIs(t, err, access.ErrInvalidToken)
	assert.NoError(t, s.Revoke(ctx, token), "twice is fine")
	assert.NoError(t, s.Revoke(ctx, ""))
}
