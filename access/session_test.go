package access_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestSessions_TTL(t *testing.T) {
	t.Parallel()
	s := access.NewSessions(nil, nil, 12*time.Hour)
	assert.Equal(t, 12*time.Hour, s.TTL())
}
