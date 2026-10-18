package access_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestNewSessions(t *testing.T) {
	t.Parallel()
	s := access.NewSessions("secret", time.Hour)
	assert.NotNil(t, s)
	assert.Equal(t, time.Hour, s.TTL())

	// Without a secret, each process signs with its own.
	a, b := access.NewSessions("", time.Hour), access.NewSessions("", time.Hour)
	_, err := b.Verify(a.Issue("x"))
	assert.ErrorIs(t, err, access.ErrInvalidToken)
}

func TestSessions_Issue(t *testing.T) {
	t.Parallel()
	s := access.NewSessions("secret", time.Hour)
	token := s.Issue("ops")
	name, err := s.Verify(token)
	assert.NoError(t, err)
	assert.Equal(t, "ops", name)
}

func TestSessions_Verify(t *testing.T) {
	t.Parallel()
	s := access.NewSessions("secret", time.Hour)
	token := s.Issue("ops")
	_, err := access.NewSessions("other", time.Hour).Verify(token)
	assert.ErrorIs(t, err, access.ErrInvalidToken)
	_, err = s.Verify("a.b")
	assert.ErrorIs(t, err, access.ErrInvalidToken)
	_, err = s.Verify("")
	assert.ErrorIs(t, err, access.ErrInvalidToken)

	s.Now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	_, err = s.Verify(token)
	assert.ErrorIs(t, err, access.ErrInvalidToken, "expired")
}

func TestSessions_TTL(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 12*time.Hour, access.NewSessions("x", 12*time.Hour).TTL())
}
