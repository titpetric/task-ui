package access

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"
)

// TokenPrefix starts every token, so a reader spots one in a
// configuration file.
const TokenPrefix = "tui_"

// MCPToken names the token a user's settings enable for MCP, one per
// user, and MCPTokenTTL is how long it lasts before a rotation.
const (
	MCPToken    = "mcp"
	MCPTokenTTL = 30 * 24 * time.Hour
)

// ErrTokenExpired reports a token past its expiry: it signs in nobody
// until its user rotates it.
var ErrTokenExpired = errors.New("token expired, rotate it on your user settings")

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

// HashToken returns the hash storage keeps a token under and looks it up
// by. A token is random enough that a plain sha256 does; a password needs
// bcrypt.
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
