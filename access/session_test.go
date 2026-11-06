package access_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

// The policy is read back as it was given, and a Touch left out is a minute.
func TestSessions_Policy(t *testing.T) {
	t.Parallel()
	s := access.NewSessions(nil, nil, access.SessionPolicy{TTL: 12 * time.Hour, Max: 30 * 24 * time.Hour})
	assert.Equal(t, 12*time.Hour, s.Policy().TTL)
	assert.Equal(t, 30*24*time.Hour, s.Policy().Max)
	assert.Equal(t, time.Minute, s.Policy().Touch)

	s = access.NewSessions(nil, nil, access.SessionPolicy{TTL: time.Hour, Touch: 5 * time.Second})
	assert.Equal(t, 5*time.Second, s.Policy().Touch)
}
