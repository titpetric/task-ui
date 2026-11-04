package view

import (
	"time"

	"github.com/titpetric/task-ui/model"
)

// expiresSoon is how close to its expiry a token shows in amber.
const expiresSoon = 7 * 24 * time.Hour

// MCPToken is one of a user's MCP tokens as a row of the table on their
// page: its ID to delete it by, the label that tells it from the others,
// and its dates. It never holds the token or its hash.
type MCPToken struct {
	ID        string
	Name      string
	CreatedAt *time.Time
	ExpiresAt *time.Time
	UsedAt    *time.Time
}

// Expired reports a token past its expiry by a clock.
func (t MCPToken) Expired(now time.Time) bool {
	return t.ExpiresAt != nil && !now.Before(*t.ExpiresAt)
}

// Tone names the badge class of the expiry: bad once expired, signal
// within a week of it, plain before.
func (t MCPToken) Tone(now time.Time) string {
	switch {
	case t.ExpiresAt == nil:
		return ""
	case t.Expired(now):
		return "bad"
	case t.ExpiresAt.Sub(now) <= expiresSoon:
		return "signal"
	}
	return ""
}

// MCPAccess is a user's MCP tokens as their settings show them: the rows
// of the table, newest first, and what the viewer may do with them.
type MCPAccess struct {
	Tokens []MCPToken
	// CanEnable issues one, with user/edit; CanManage issues and deletes,
	// with user/edit or as its own user once the user holds one.
	CanEnable bool
	CanManage bool
	// Agent says the user is an MCP agent user, the flag a token needs.
	Agent bool
}

// NewMCPAccess returns the access of a user's token rows.
func NewMCPAccess(tokens []model.UserAuthToken) MCPAccess {
	out := MCPAccess{Tokens: make([]MCPToken, 0, len(tokens))}
	for _, token := range tokens {
		out.Tokens = append(out.Tokens, MCPToken{
			ID: token.ID, Name: token.Name, CreatedAt: token.CreatedAt, ExpiresAt: token.ExpiresAt, UsedAt: token.UsedAt,
		})
	}
	return out
}

// Enabled reports the user holding a token, which is what MCP being on
// for them means: there is no flag beside the rows to go out of step with
// them.
func (a MCPAccess) Enabled() bool {
	return len(a.Tokens) > 0
}
