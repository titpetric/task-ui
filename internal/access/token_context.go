package access

import (
	"context"
)

type tokenKey struct{}

type viaKey struct{}

// TokenFrom returns the ID of the token that signed a request in, "" for a session.
func TokenFrom(ctx context.Context) string {
	id, _ := ctx.Value(tokenKey{}).(string)
	return id
}

// WithToken returns a context naming the token that signed a request in.
func WithToken(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, tokenKey{}, id)
}

// ViaMCP is the way of a request that an MCP tool makes, in process.
const ViaMCP = "mcp"

// ViaAPI is the way of a request a client sends to the API.
const ViaAPI = "api"

// Via returns how a request reached the API: ViaMCP for a tool's call,
// else ViaAPI.
func Via(ctx context.Context) string {
	if via, _ := ctx.Value(viaKey{}).(string); via != "" {
		return via
	}
	return ViaAPI
}

// WithVia returns a context marking how its requests reach the API.
func WithVia(ctx context.Context, via string) context.Context {
	return context.WithValue(ctx, viaKey{}, via)
}
