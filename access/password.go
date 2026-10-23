package access

import (
	"golang.org/x/crypto/bcrypt"
)

// dummyHash stands in when a login names no user, so a missing
// user costs the same time as a wrong password.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("task-ui"), bcrypt.DefaultCost)

// HashPassword returns the bcrypt hash of a password, as the password file and user_auth_password store it.
func HashPassword(password string) (string, error) {
	out, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(out), err
}

// CheckHash reports whether a password matches a bcrypt hash; an empty hash never matches.
// An empty hash takes as long as a real comparison, so a missing user costs the same time as a wrong password.
func CheckHash(hash, password string) bool {
	if hash == "" {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Hashed reports whether text is a bcrypt hash, as a configured password must be.
func Hashed(text string) bool {
	return len(text) > 4 && text[0] == '$' && text[1] == '2'
}
