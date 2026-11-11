package forgejo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Sign returns the signature of a hook's body under a secret, X-Forgejo-Signature: the HMAC-SHA256 of the body in hex.
func Sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify reports whether a signature is the body's under the secret,
// compared in constant time. An empty secret verifies nothing.
func Verify(secret string, body []byte, signature string) bool {
	if secret == "" || signature == "" {
		return false
	}
	return hmac.Equal([]byte(Sign(secret, body)), []byte(signature))
}
