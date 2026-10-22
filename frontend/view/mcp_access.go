package view

import (
	"time"

	"github.com/titpetric/task-ui/model"
)

// expiresSoon is how close to its expiry a token shows in amber.
const expiresSoon = 7 * 24 * time.Hour

// MCPAccess is a user's MCP token as their settings show it: whether it
// is on, its dates, and what the viewer may do with it. It never holds
// the token or its hash.
type MCPAccess struct {
	Enabled   bool
	CreatedAt *time.Time
	ExpiresAt *time.Time
	UsedAt    *time.Time
	// CanEnable turns it on, with user/edit; CanManage rotates and
	// disables it, with user/edit or as its own user once on.
	CanEnable bool
	CanManage bool
}

// NewMCPAccess returns the access of a token row; nil is off.
func NewMCPAccess(token *model.UserAuthToken) MCPAccess {
	if token == nil {
		return MCPAccess{}
	}
	return MCPAccess{Enabled: true, CreatedAt: token.CreatedAt, ExpiresAt: token.ExpiresAt, UsedAt: token.UsedAt}
}

// Expired reports a token past its expiry by a clock.
func (a MCPAccess) Expired(now time.Time) bool {
	return a.ExpiresAt != nil && !now.Before(*a.ExpiresAt)
}

// Tone names the badge class of the expiry: bad once expired, signal
// within a week of it, plain before.
func (a MCPAccess) Tone(now time.Time) string {
	switch {
	case a.ExpiresAt == nil:
		return ""
	case a.Expired(now):
		return "bad"
	case a.ExpiresAt.Sub(now) <= expiresSoon:
		return "signal"
	}
	return ""
}
