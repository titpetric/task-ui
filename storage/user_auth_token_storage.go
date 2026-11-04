package storage

import (
	"context"
	"errors"
	"time"

	"github.com/titpetric/task-ui/model"
)

// UserAuthTokenStorage keeps the tokens clients sign requests with, as
// hashes, each under a name of its user's choosing.
//
// A name is unique per user and kind, so a user holds as many tokens of a
// kind as they name: Set replaces the one of a name, which is what
// `task-ui token` means by a name, and Add writes one more, which is what
// a user's MCP access does.
type UserAuthTokenStorage struct {
	db DB
}

// NewUserAuthTokenStorage returns the storage over a pool.
func NewUserAuthTokenStorage(db DB) *UserAuthTokenStorage {
	return &UserAuthTokenStorage{db: db}
}

// Set stores a CLI token that never expires for a user under a name, replacing the one it had.
func (s *UserAuthTokenStorage) Set(ctx context.Context, userID, name, hash string) error {
	if err := s.Delete(ctx, userID, model.TokenCLI, name); err != nil {
		return err
	}
	_, err := s.Add(ctx, userID, model.TokenCLI, name, hash, nil)
	return err
}

// Add stores one more token of a kind for a user and returns its row; it replaces nothing, so a name the user already holds for that kind is an error.
func (s *UserAuthTokenStorage) Add(ctx context.Context, userID, kind, name, hash string, expires *time.Time) (*model.UserAuthToken, error) {
	if userID == "" || name == "" || hash == "" {
		return nil, errors.New("a token needs a user, a name and a hash")
	}
	if expires != nil {
		at := expires.UTC().Truncate(time.Millisecond)
		expires = &at
	}
	row := &model.UserAuthToken{ID: model.NewID(), UserID: userID, Kind: kind, Name: name, Hash: hash, CreatedAt: stamp(), ExpiresAt: expires}
	row.UpdatedAt = row.CreatedAt
	return row, client(s.db).Insert(ctx, model.UserAuthTokenTable, row)
}

// Get returns a user's token of a kind by name, or ErrNotFound.
func (s *UserAuthTokenStorage) Get(ctx context.Context, userID, kind, name string) (*model.UserAuthToken, error) {
	token, err := client(s.db).Get[model.UserAuthToken](ctx, "SELECT * FROM user_auth_token WHERE user_id = ? AND kind = ? AND name = ?", userID, kind, name)
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

// List returns a user's tokens of every kind, by name.
func (s *UserAuthTokenStorage) List(ctx context.Context, userID string) ([]model.UserAuthToken, error) {
	return client(s.db).Select[model.UserAuthToken](ctx, "SELECT * FROM user_auth_token WHERE user_id = ? ORDER BY name", userID)
}

// ListKind returns a user's tokens of one kind, newest first: the order their page lists them in, because the one they just issued is the one they are looking for.
func (s *UserAuthTokenStorage) ListKind(ctx context.Context, userID, kind string) ([]model.UserAuthToken, error) {
	return client(s.db).Select[model.UserAuthToken](ctx, "SELECT * FROM user_auth_token WHERE user_id = ? AND kind = ? ORDER BY created_at DESC, name", userID, kind)
}

// Delete removes a user's token of a kind by name.
func (s *UserAuthTokenStorage) Delete(ctx context.Context, userID, kind, name string) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_auth_token WHERE user_id = ? AND kind = ? AND name = ?", userID, kind, name)
}

// DeleteID removes one of a user's tokens by its ID; an ID another user holds removes nothing, so a route needs no second read to scope it.
func (s *UserAuthTokenStorage) DeleteID(ctx context.Context, userID, id string) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_auth_token WHERE user_id = ? AND id = ?", userID, id)
}

// DeleteByUser removes every token of a user.
func (s *UserAuthTokenStorage) DeleteByUser(ctx context.Context, userID string) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_auth_token WHERE user_id = ?", userID)
}

// Expired returns the tokens that stopped signing in by now, oldest expiry first.
func (s *UserAuthTokenStorage) Expired(ctx context.Context, now time.Time) ([]model.UserAuthToken, error) {
	return client(s.db).Select[model.UserAuthToken](ctx, "SELECT * FROM user_auth_token WHERE expires_at IS NOT NULL AND expires_at <= ? ORDER BY expires_at ASC", now.UTC())
}

// DeleteExpired removes the tokens that stopped signing in by now.
func (s *UserAuthTokenStorage) DeleteExpired(ctx context.Context, now time.Time) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_auth_token WHERE expires_at IS NOT NULL AND expires_at <= ?", now.UTC())
}

// DeleteKind removes every token of one kind of a user: model.TokenMCP leaves the tokens of `task-ui token` alone.
func (s *UserAuthTokenStorage) DeleteKind(ctx context.Context, userID, kind string) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_auth_token WHERE user_id = ? AND kind = ?", userID, kind)
}
