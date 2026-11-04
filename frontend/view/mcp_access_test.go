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
func TestMCPToken(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.October, 22, 12, 0, 0, 0, time.UTC)
	assert.False(t, view.NewMCPAccess(nil).Enabled())

	expires := now.Add(20 * 24 * time.Hour)
	access := view.NewMCPAccess([]model.UserAuthToken{{CreatedAt: &now, ExpiresAt: &expires}})
	assert.True(t, access.Enabled())

	token := access.Tokens[0]
	assert.Equal(t, "", token.Tone(now))
	assert.Equal(t, "signal", token.Tone(expires.Add(-6*24*time.Hour)))
	assert.False(t, token.Expired(expires.Add(-time.Second)))
	assert.True(t, token.Expired(expires))
	assert.Equal(t, "bad", token.Tone(expires))

	forever := view.NewMCPAccess([]model.UserAuthToken{{CreatedAt: &now}}).Tokens[0]
	assert.False(t, forever.Expired(now.Add(1000*24*time.Hour)))
	assert.Equal(t, "", forever.Tone(now))
}

// Every row keeps its ID, because that is what the Delete of a row calls,
// and the order storage answered in, newest first, because that is the
// order the table reads in.
func TestMCPAccess(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.October, 22, 12, 0, 0, 0, time.UTC)
	older := now.Add(-48 * time.Hour)

	access := view.NewMCPAccess([]model.UserAuthToken{
		{ID: "02", Name: "mcp-2", CreatedAt: &now, UsedAt: &now},
		{ID: "01", Name: "mcp", CreatedAt: &older},
	})
	assert.Equal(t, 2, len(access.Tokens))
	assert.Equal(t, "02", access.Tokens[0].ID)
	assert.Equal(t, "mcp-2", access.Tokens[0].Name)
	assert.Equal(t, &now, access.Tokens[0].UsedAt)
	assert.Equal(t, "mcp", access.Tokens[1].Name)
	assert.Equal(t, (*time.Time)(nil), access.Tokens[1].UsedAt, "a token nobody signed a request with")

	// No row is no access, and the flags are the caller's to set: nothing
	// about a list of tokens says who may delete one.
	off := view.NewMCPAccess(nil)
	assert.False(t, off.Enabled())
	assert.False(t, off.CanManage)
	assert.False(t, off.Agent)
}
