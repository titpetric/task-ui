package access

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// ErrInvalidToken is returned for a session token that does not verify,
// or has expired.
var ErrInvalidToken = errors.New("invalid session")

// Sessions signs and verifies the session cookie: an HS256 JWT holding the
// username. The role is looked up from the configuration on every request,
// so a user removed from the password file loses access on their next click.
type Sessions struct {
	secret []byte
	ttl    time.Duration
	// Now is the clock; tests move it.
	Now func() time.Time
}

// NewSessions builds the signer. An empty secret generates one per
// process, so sessions do not survive a restart.
func NewSessions(secret string, ttl time.Duration) *Sessions {
	key := []byte(secret)
	if len(key) == 0 {
		key = make([]byte, 32)
		_, _ = rand.Read(key)
	}
	return &Sessions{secret: key, ttl: ttl, Now: time.Now}
}

func segment(v any) string {
	data, _ := json.Marshal(v)
	return base64.RawURLEncoding.EncodeToString(data)
}

func (s *Sessions) sign(input string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Issue returns a session token for a user.
func (s *Sessions) Issue(username string) string {
	now := s.Now()
	signed := segment(map[string]string{"alg": "HS256", "typ": "JWT"}) + "." +
		segment(map[string]any{"sub": username, "iat": now.Unix(), "exp": now.Add(s.ttl).Unix()})
	return signed + "." + s.sign(signed)
}

// Verify returns the username a token was issued to.
func (s *Sessions) Verify(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || !hmac.Equal([]byte(s.sign(parts[0]+"."+parts[1])), []byte(parts[2])) {
		return "", ErrInvalidToken
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", ErrInvalidToken
	}
	var claims struct {
		Sub string `json:"sub"`
		Exp int64  `json:"exp"`
	}
	if err := json.Unmarshal(data, &claims); err != nil || claims.Sub == "" {
		return "", ErrInvalidToken
	}
	if claims.Exp != 0 && s.Now().Unix() >= claims.Exp {
		return "", ErrInvalidToken
	}
	return claims.Sub, nil
}

// TTL is how long an issued session stays valid.
func (s *Sessions) TTL() time.Duration {
	return s.ttl
}
