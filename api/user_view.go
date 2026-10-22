package api

import (
	"context"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
)

// UserView is a user as the API shows one: the row, and whether the
// user has MCP access, never what they sign in with.
type UserView struct {
	model.User
	MCP UserMCP `json:"mcp"`
}

// userView returns the view of a user.
func (h *Handlers) userView(ctx context.Context, user *model.User) UserView {
	view := UserView{User: *user}
	if h.opts.Tokens == nil {
		return view
	}
	if token, err := h.opts.Tokens.Get(ctx, user.ID, access.MCPToken); err == nil {
		view.MCP = UserMCP{Enabled: true, CreatedAt: token.CreatedAt, ExpiresAt: token.ExpiresAt, UsedAt: token.UsedAt}
	}
	return view
}
