package access

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

// TokenPrefix starts every token, marking one in a configuration file.
const TokenPrefix = "tui_"

// MCPTokenName labels a user's first MCP token and MCPTokenTTL sets the lifetime of every one.
// model.TokenMCP marks the kind. Storage keys on kind and label; a user may name any number of them.
const (
	MCPTokenName = "mcp"
	MCPTokenTTL  = 30 * 24 * time.Hour
)

// ErrTokenExpired reports a token past its expiry: it signs in nobody
// until its user issues another.
var ErrTokenExpired = errors.New("the token has expired; issue another under your MCP access")

// NewToken returns a new token for a client, 256 random bits, and its
// hash. The caller shows the token once and keeps only the hash.
func NewToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	token := TokenPrefix + hex.EncodeToString(buf)
	return token, HashToken(token), nil
}

// TokenName returns want when taken does not hold it, else want with -2, -3 and so on appended until one is free.
// An empty want starts from MCPTokenName: mcp, mcp-2, mcp-3.
// A label is unique among a user's tokens of one kind, and a reader of the list has to tell two of them apart.
func TokenName(want string, taken []string) string {
	stem := strings.TrimSpace(want)
	if stem == "" {
		stem = MCPTokenName
	}
	for name, n := stem, 2; ; n++ {
		if !slices.Contains(taken, name) {
			return name
		}
		name = stem + "-" + strconv.Itoa(n)
	}
}

// HashToken returns the sha256 hash storage keeps a token under: a token is random enough to need no bcrypt.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// BearerToken returns the token of an Authorization: Bearer header, or "".
func BearerToken(r *http.Request) string {
	scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}
