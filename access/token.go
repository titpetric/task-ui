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

// TokenPrefix starts every token, so a reader spots one in a
// configuration file.
const TokenPrefix = "tui_"

// MCPTokenName is the label the first MCP token of a user takes, and
// MCPTokenTTL is how long any of them lasts. The kind that marks one is
// model.TokenMCP: storage keys on the kind and the label is a user's own,
// so a user holds as many as they care to name.
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

// TokenName returns the label a new token takes: the one asked for where
// nothing of taken holds it, and otherwise MCPTokenName with a number
// after it, counting up to the first that is free.
//
// A label is unique among a user's tokens of one kind, so issuing one
// without naming it has to pick a name nobody took, and a reader of the
// list has to be able to tell which token a client holds. mcp, mcp-2 and
// mcp-3 read as what they are and keep the name an instance already gave
// the single token MCP access used to issue.
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
