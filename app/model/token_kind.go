package model

// TokenKind is what issued a token, the kind column of
// phpfront/app/modules/mcp/schema/user_auth_token_kind.up.sql.
//
// It is what tells a user's tokens apart on their page, because nothing in
// signing a request reads it: the Guard finds a token by its hash alone, so
// every kind signs in the same way and for as long as its expiry says.
type TokenKind = string

// The kinds of token a client signs requests with.
const (
	// TokenCLI is a token `task-ui token` wrote: it never expires, and no
	// page lists it or deletes it.
	TokenCLI TokenKind = ""
	// TokenMCP is a token a user's MCP access issued: it expires, and
	// their page lists it, issues another and deletes one.
	TokenMCP TokenKind = "mcp"
)
