package access

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/storage"
)

// ErrInvalidToken reports a session token that names no session or has expired.
var ErrInvalidToken = errors.New("invalid session")

// Sessions issues and resolves the session cookie: an opaque random
// token the server looks up by its hash. The database holds the hash
// and the user it names. A user removed or given a new password loses
// their sessions.
type Sessions struct {
	Store  *storage.UserSessionStorage
	Users  *storage.UserStorage
	policy SessionPolicy
	// Now is the clock; tests move it.
	Now func() time.Time
}

// SessionPolicy sets a session's lifetime and how often a use is written.
//
// TTL is the window a session in use keeps ahead of it. Max is the ceiling
// from the sign in, which no extension passes; a Max at or below TTL pins
// the session to one TTL from the sign in. Touch is the staleness a use
// reaches before a request records it: a click inside Touch writes nothing.
type SessionPolicy struct {
	TTL   time.Duration
	Max   time.Duration
	Touch time.Duration
}

// NewSessions returns the sessions over their storage and the users. A
// policy with no Touch records a use at most once a minute.
func NewSessions(store *storage.UserSessionStorage, users *storage.UserStorage, policy SessionPolicy) *Sessions {
	if policy.Touch <= 0 {
		policy.Touch = time.Minute
	}
	return &Sessions{Store: store, Users: users, policy: policy, Now: time.Now}
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
	if _, err := s.Store.Create(ctx, user.ID, HashToken(token), now.Add(s.policy.TTL)); err != nil {
		return "", err
	}
	return token, nil
}

// Resolve returns the user of a session token and the expiry the use moved it to.
// A zero time is a session this call left alone.
func (s *Sessions) Resolve(ctx context.Context, token string) (*model.User, time.Time, error) {
	if token == "" {
		return nil, time.Time{}, ErrInvalidToken
	}
	row, err := s.Store.Find(ctx, HashToken(token))
	if errors.Is(err, storage.ErrNotFound) {
		return nil, time.Time{}, ErrInvalidToken
	} else if err != nil {
		return nil, time.Time{}, err
	}
	now := s.Now()
	if row.ExpiresAt == nil || !now.Before(*row.ExpiresAt) {
		return nil, time.Time{}, ErrInvalidToken
	}
	user, err := s.Users.Get(ctx, row.UserID)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, time.Time{}, ErrInvalidToken
	} else if err != nil {
		return nil, time.Time{}, err
	}
	if row.UsedAt != nil && now.Sub(*row.UsedAt) < s.policy.Touch {
		return user, time.Time{}, nil
	}
	until := s.until(row, now)
	if until.IsZero() {
		_ = s.Store.Touch(ctx, row.ID)
		return user, time.Time{}, nil
	}
	if err := s.Store.Extend(ctx, row.ID, until); err != nil {
		return nil, time.Time{}, err
	}
	return user, until, nil
}

// until is the expiry a session in use moves to, one TTL ahead and no
// further than Max from the sign in; a zero time leaves the row alone.
func (s *Sessions) until(row *model.UserSession, now time.Time) time.Time {
	if s.policy.Max <= s.policy.TTL || row.CreatedAt == nil {
		return time.Time{}
	}
	until := now.Add(s.policy.TTL)
	if ceiling := row.CreatedAt.Add(s.policy.Max); until.After(ceiling) {
		until = ceiling
	}
	if !until.After(*row.ExpiresAt) {
		return time.Time{}
	}
	return until
}

// Revoke ends the session of a token. A token that names none is fine.
func (s *Sessions) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.Store.Delete(ctx, HashToken(token))
}

// Policy returns the policy NewSessions took, with Touch defaulted.
func (s *Sessions) Policy() SessionPolicy {
	return s.policy
}
