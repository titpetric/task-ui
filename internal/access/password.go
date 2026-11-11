package access

import (
	"golang.org/x/crypto/bcrypt"
)

// dummyHash is the hash CheckHash compares against when no user matched.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("task-ui"), bcrypt.DefaultCost)

// HashPassword returns the bcrypt hash of a password, as the password file and user_auth_password store it.
func HashPassword(password string) (string, error) {
	out, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(out), err
}

// CheckHash reports whether a password matches a bcrypt hash. An empty hash never matches and still takes the time of a real comparison: a missing user and a wrong password cost the same.
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
