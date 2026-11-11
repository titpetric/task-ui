package access_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/internal/assert"
)

// The policy reads back unchanged, and a Touch left out defaults to a minute.
func TestSessions_Policy(t *testing.T) {
	t.Parallel()
	s := access.NewSessions(nil, nil, access.SessionPolicy{TTL: 12 * time.Hour, Max: 30 * 24 * time.Hour})
	assert.Equal(t, 12*time.Hour, s.Policy().TTL)
	assert.Equal(t, 30*24*time.Hour, s.Policy().Max)
	assert.Equal(t, time.Minute, s.Policy().Touch)

	s = access.NewSessions(nil, nil, access.SessionPolicy{TTL: time.Hour, Touch: 5 * time.Second})
	assert.Equal(t, 5*time.Second, s.Policy().Touch)
}
