package storage

import (
	"context"
	"errors"
	"time"

	"github.com/titpetric/task-ui/model"
)

// UserSessionStorage keeps the browser sessions, each by the hash of
// its cookie, so the database holds nothing a browser could present.
type UserSessionStorage struct {
	db DB
}

// NewUserSessionStorage returns the storage over a pool.
func NewUserSessionStorage(db DB) *UserSessionStorage {
	return &UserSessionStorage{db: db}
}

// Create stores a session for a user under a hash, valid until expires.
func (s *UserSessionStorage) Create(ctx context.Context, userID, hash string, expires time.Time) (*model.UserSession, error) {
	if userID == "" || hash == "" {
		return nil, errors.New("a session needs a user and a hash")
	}
	expires = expires.UTC().Truncate(time.Millisecond)
	row := &model.UserSession{ID: model.NewID(), UserID: userID, Hash: hash, ExpiresAt: &expires, CreatedAt: stamp()}
	row.UpdatedAt = row.CreatedAt
	return row, client(s.db).Insert(ctx, model.UserSessionTable, row)
}

// Find returns the session with a hash, or ErrNotFound, without looking at the expiry: the caller holds the clock.
func (s *UserSessionStorage) Find(ctx context.Context, hash string) (*model.UserSession, error) {
	session, err := client(s.db).Get[model.UserSession](ctx, "SELECT * FROM user_session WHERE hash = ?", hash)
	return session, notFound(err)
}

// Touch records a session's use.
func (s *UserSessionStorage) Touch(ctx context.Context, id string) error {
	now := stamp()
	return client(s.db).Exec(ctx, "UPDATE user_session SET used_at = ?, updated_at = ? WHERE id = ?", now, now, id)
}

// Delete removes the session with a hash.
func (s *UserSessionStorage) Delete(ctx context.Context, hash string) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_session WHERE hash = ?", hash)
}

// DeleteByUser removes every session of a user.
func (s *UserSessionStorage) DeleteByUser(ctx context.Context, userID string) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_session WHERE user_id = ?", userID)
}

// DeleteExpired removes the sessions that expired by now.
func (s *UserSessionStorage) DeleteExpired(ctx context.Context, now time.Time) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_session WHERE expires_at <= ?", now.UTC())
}
