//go:build integration

package access_test

import (
	"context"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/storage"
)

// newSessions opens a private in-memory database with the schema applied
// and returns the sessions over it, the user they sign in, and the clock.
func newSessions(t *testing.T, ttl, max time.Duration) (*access.Sessions, *model.User, *time.Time) {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:", app.Migrations())
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	users := storage.NewUserStorage(db)
	user := &model.User{Name: "reader", Role: model.RoleViewer}
	assert.NoError(t, users.Create(context.Background(), user))

	now := time.Now().UTC().Truncate(time.Millisecond)
	sessions := access.NewSessions(storage.NewUserSessionStorage(db), users, access.SessionPolicy{TTL: ttl, Max: max})
	sessions.Now = func() time.Time { return now }
	return sessions, user, &now
}

// A session in use keeps one ttl ahead of the clock, and the expiry it
// moves to is what Resolve reports so the cookie follows the row.
func TestSessions_Resolve(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sessions, user, clock := newSessions(t, time.Hour, 30*24*time.Hour)

	token, err := sessions.Issue(ctx, user)
	assert.NoError(t, err)

	got, until, err := sessions.Resolve(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, got.ID)
	assert.True(t, until.IsZero(), "the first resolve is inside the touch window")

	*clock = clock.Add(2 * time.Minute)
	got, until, err = sessions.Resolve(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, got.ID)
	assert.Equal(t, clock.Add(time.Hour), until)

	*clock = clock.Add(59 * time.Minute)
	_, _, err = sessions.Resolve(ctx, token)
	assert.NoError(t, err, "an hour past the sign in, the extension carried it")

	*clock = clock.Add(2 * time.Hour)
	_, _, err = sessions.Resolve(ctx, token)
	assert.ErrorIs(t, err, access.ErrInvalidToken, "two hours idle is past the window")
}

// The ceiling is the sign in plus max: extensions stop there, and the
// session ends however much it is used.
func TestSessions_ResolveCeiling(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sessions, user, clock := newSessions(t, time.Hour, 90*time.Minute)
	opened := *clock

	token, err := sessions.Issue(ctx, user)
	assert.NoError(t, err)

	*clock = clock.Add(30 * time.Minute)
	_, until, err := sessions.Resolve(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, opened.Add(90*time.Minute), until, "the ceiling cuts the hour short")

	*clock = clock.Add(30 * time.Minute)
	_, until, err = sessions.Resolve(ctx, token)
	assert.NoError(t, err)
	assert.True(t, until.IsZero(), "at the ceiling nothing moves")

	*clock = clock.Add(31 * time.Minute)
	_, _, err = sessions.Resolve(ctx, token)
	assert.ErrorIs(t, err, access.ErrInvalidToken, "the ceiling ends the session")
}

// A max at or below the ttl extends nothing: the session lives one ttl
// from the sign in, as it did before the window slid.
func TestSessions_ResolveFixed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sessions, user, clock := newSessions(t, time.Hour, 0)

	token, err := sessions.Issue(ctx, user)
	assert.NoError(t, err)

	*clock = clock.Add(30 * time.Minute)
	_, until, err := sessions.Resolve(ctx, token)
	assert.NoError(t, err)
	assert.True(t, until.IsZero())

	*clock = clock.Add(31 * time.Minute)
	_, _, err = sessions.Resolve(ctx, token)
	assert.ErrorIs(t, err, access.ErrInvalidToken)
}

// Issue opens a session a token resolves, and Revoke ends it.
func TestSessions_Issue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sessions, user, _ := newSessions(t, time.Hour, 0)

	token, err := sessions.Issue(ctx, user)
	assert.NoError(t, err)
	assert.NotEqual(t, "", token)

	other, err := sessions.Issue(ctx, user)
	assert.NoError(t, err)
	assert.NotEqual(t, token, other, "each sign in is its own session")

	_, _, err = sessions.Resolve(ctx, token)
	assert.NoError(t, err)
}

// Revoke ends one session and leaves the user's others alone; a token
// that names none is no error.
func TestSessions_Revoke(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sessions, user, _ := newSessions(t, time.Hour, 0)

	token, err := sessions.Issue(ctx, user)
	assert.NoError(t, err)
	other, err := sessions.Issue(ctx, user)
	assert.NoError(t, err)

	assert.NoError(t, sessions.Revoke(ctx, ""))
	assert.NoError(t, sessions.Revoke(ctx, "names-nothing"))
	assert.NoError(t, sessions.Revoke(ctx, token))

	_, _, err = sessions.Resolve(ctx, token)
	assert.ErrorIs(t, err, access.ErrInvalidToken)
	_, _, err = sessions.Resolve(ctx, other)
	assert.NoError(t, err)
}
