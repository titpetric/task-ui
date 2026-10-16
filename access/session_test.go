package access

import (
	"errors"
	"testing"
	"time"
)

func TestSessions(t *testing.T) {
	s := NewSessions("secret", time.Hour)
	if s.TTL() != time.Hour {
		t.Fatal("ttl")
	}
	token := s.Issue("ops")
	if name, err := s.Verify(token); err != nil || name != "ops" {
		t.Fatalf("verify: %q %v", name, err)
	}
	if _, err := NewSessions("other", time.Hour).Verify(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("token verified with another secret")
	}
	if _, err := s.Verify("a.b"); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("malformed token verified")
	}
	s.Now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	if _, err := s.Verify(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatal("expired token verified")
	}

	// Without a secret, each process signs with its own.
	a, b := NewSessions("", time.Hour), NewSessions("", time.Hour)
	if _, err := b.Verify(a.Issue("x")); err == nil {
		t.Fatal("generated secrets collided")
	}
}
