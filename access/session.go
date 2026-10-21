package access

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// ErrInvalidToken reports a session token that names no session, or one
// that has expired.
var ErrInvalidToken = errors.New("invalid session")

// touchEvery is how stale a session's last use gets before a request
// records it again, so page loads do not write on every click.
const touchEvery = time.Minute

// Sessions issues and resolves the session cookie: an opaque random
// token the server looks up by its hash. The database holds the hash
// alone, and the user it names, so a user removed or given a new
// password loses their sessions with it.
type Sessions struct {
	Store *storage.UserSessionStorage
	Users *storage.UserStorage
	ttl   time.Duration
	// Now is the clock; tests move it.
	Now func() time.Time
}

// NewSessions returns the sessions over their storage and the users.
func NewSessions(store *storage.UserSessionStorage, users *storage.UserStorage, ttl time.Duration) *Sessions {
	return &Sessions{Store: store, Users: users, ttl: ttl, Now: time.Now}
}

// Issue opens a session for a user and returns its token, 256 random
// bits. It prunes the expired sessions on the way.
func (s *Sessions) Issue(ctx context.Context, user *model.User) (string, error) {
	now := s.Now()
	if err := s.Store.DeleteExpired(ctx, now); err != nil {
		return "", err
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	if _, err := s.Store.Create(ctx, user.ID, HashToken(token), now.Add(s.ttl)); err != nil {
		return "", err
	}
	return token, nil
}

// Resolve returns the user of a session token, and records the use at
// most once a minute.
func (s *Sessions) Resolve(ctx context.Context, token string) (*model.User, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}
	row, err := s.Store.Find(ctx, HashToken(token))
	if errors.Is(err, storage.ErrNotFound) {
		return nil, ErrInvalidToken
	} else if err != nil {
		return nil, err
	}
	now := s.Now()
	if row.ExpiresAt == nil || !now.Before(*row.ExpiresAt) {
		return nil, ErrInvalidToken
	}
	user, err := s.Users.Get(ctx, row.UserID)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, ErrInvalidToken
	} else if err != nil {
		return nil, err
	}
	if row.UsedAt == nil || now.Sub(*row.UsedAt) >= touchEvery {
		_ = s.Store.Touch(ctx, row.ID)
	}
	return user, nil
}

// Revoke ends the session of a token. A token that names none is fine.
func (s *Sessions) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.Store.Delete(ctx, HashToken(token))
}

// TTL is how long an issued session stays valid.
func (s *Sessions) TTL() time.Duration {
	return s.ttl
}
