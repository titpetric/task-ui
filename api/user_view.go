package api

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// UserView is a user as the API shows one: the row, with MCPAgent, the
// row's mcp flag, and the MCP tokens the user holds, never what they sign
// in with.
type UserView struct {
	model.User
	MCPAgent bool    `json:"mcp_agent"`
	MCP      UserMCP `json:"mcp"`
}

// userView returns the view of a user.
func (h *Handlers) userView(ctx context.Context, user *model.User) UserView {
	view := UserView{User: *user, MCPAgent: user.Mcp, MCP: newUserMCP(nil)}
	if h.opts.Tokens == nil {
		return view
	}
	if rows, err := h.opts.Tokens.ListKind(ctx, user.ID, model.TokenMCP); err == nil {
		view.MCP = newUserMCP(rows)
	}
	return view
}
