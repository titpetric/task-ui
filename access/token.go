package access

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

// TokenPrefix starts every token, so one is recognised in a config file.
const TokenPrefix = "tui_"

// NewToken returns a new token for a client, 256 random bits, and its
// hash. The token is shown once; only the hash is kept.
func NewToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	token := TokenPrefix + hex.EncodeToString(buf)
	return token, HashToken(token), nil
}

// HashToken returns the hash a token is kept and looked up by. A token is
// random enough that a plain sha256 does; a password needs bcrypt.
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
