package view_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The expiry reads plain, amber within a week, red once past; a token
// without one never expires.
func TestMCPAccess(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.October, 22, 12, 0, 0, 0, time.UTC)
	assert.False(t, view.NewMCPAccess(nil).Enabled)

	expires := now.Add(20 * 24 * time.Hour)
	access := view.NewMCPAccess(&model.UserAuthToken{CreatedAt: &now, ExpiresAt: &expires})
	assert.True(t, access.Enabled)
	assert.Equal(t, "", access.Tone(now))
	assert.Equal(t, "signal", access.Tone(expires.Add(-6*24*time.Hour)))
	assert.False(t, access.Expired(expires.Add(-time.Second)))
	assert.True(t, access.Expired(expires))
	assert.Equal(t, "bad", access.Tone(expires))

	forever := view.NewMCPAccess(&model.UserAuthToken{CreatedAt: &now})
	assert.False(t, forever.Expired(now.Add(1000*24*time.Hour)))
	assert.Equal(t, "", forever.Tone(now))
}
