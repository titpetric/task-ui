package mcp

import (
	"context"
	"encoding/json"
)

// Tool is a call a client can make: its name, what it does, the shape
// of its arguments, and the call.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema Schema `json:"inputSchema"`

	call func(ctx context.Context, c *client, args json.RawMessage) (string, error)
}
