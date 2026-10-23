package storage

import (
	"context"
	"errors"
	"time"

	"github.com/titpetric/task-ui/model"
)

// UserAuthTokenStorage keeps the tokens clients sign requests with, as
// hashes, each under a name of its user's choosing.
type UserAuthTokenStorage struct {
	db DB
}

// NewUserAuthTokenStorage returns the storage over a pool.
func NewUserAuthTokenStorage(db DB) *UserAuthTokenStorage {
	return &UserAuthTokenStorage{db: db}
}

// Set stores a token that never expires for a user under a name, in place of the one it had.
func (s *UserAuthTokenStorage) Set(ctx context.Context, userID, name, hash string) error {
	_, err := s.set(ctx, userID, name, hash, nil)
	return err
}

// SetUntil stores a token like Set, one that stops signing in at expires, and returns its row.
func (s *UserAuthTokenStorage) SetUntil(ctx context.Context, userID, name, hash string, expires time.Time) (*model.UserAuthToken, error) {
	expires = expires.UTC().Truncate(time.Millisecond)
	return s.set(ctx, userID, name, hash, &expires)
}

func (s *UserAuthTokenStorage) set(ctx context.Context, userID, name, hash string, expires *time.Time) (*model.UserAuthToken, error) {
	if userID == "" || name == "" || hash == "" {
		return nil, errors.New("a token needs a user, a name and a hash")
	}
	if err := s.Delete(ctx, userID, name); err != nil {
		return nil, err
	}
	row := &model.UserAuthToken{ID: model.NewID(), UserID: userID, Name: name, Hash: hash, CreatedAt: stamp(), ExpiresAt: expires}
	row.UpdatedAt = row.CreatedAt
	return row, client(s.db).Insert(ctx, model.UserAuthTokenTable, row)
}

// Get returns a user's token by name, or ErrNotFound.
func (s *UserAuthTokenStorage) Get(ctx context.Context, userID, name string) (*model.UserAuthToken, error) {
	token, err := client(s.db).Get[model.UserAuthToken](ctx, "SELECT * FROM user_auth_token WHERE user_id = ? AND name = ?", userID, name)
	return token, notFound(err)
}

// Find returns the token with a hash, or ErrNotFound.
func (s *UserAuthTokenStorage) Find(ctx context.Context, hash string) (*model.UserAuthToken, error) {
	token, err := client(s.db).Get[model.UserAuthToken](ctx, "SELECT * FROM user_auth_token WHERE hash = ?", hash)
	return token, notFound(err)
}

// Touch records a token's use.
func (s *UserAuthTokenStorage) Touch(ctx context.Context, id string) error {
	now := stamp()
	return client(s.db).Exec(ctx, "UPDATE user_auth_token SET used_at = ?, updated_at = ? WHERE id = ?", now, now, id)
}

// List returns a user's tokens by name.
func (s *UserAuthTokenStorage) List(ctx context.Context, userID string) ([]model.UserAuthToken, error) {
	return client(s.db).Select[model.UserAuthToken](ctx, "SELECT * FROM user_auth_token WHERE user_id = ? ORDER BY name", userID)
}

// Delete removes a user's token by name.
func (s *UserAuthTokenStorage) Delete(ctx context.Context, userID, name string) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_auth_token WHERE user_id = ? AND name = ?", userID, name)
}

// DeleteByUser removes every token of a user.
func (s *UserAuthTokenStorage) DeleteByUser(ctx context.Context, userID string) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_auth_token WHERE user_id = ?", userID)
}

// DeleteMCP removes the expiring MCP tokens of a user; the tokens of `task-ui token` stay.
func (s *UserAuthTokenStorage) DeleteMCP(ctx context.Context, userID string) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_auth_token WHERE user_id = ? AND expires_at IS NOT NULL", userID)
}
