package access

import (
	"golang.org/x/crypto/bcrypt"
)

// dummyHash is compared against when a login names no user, so a missing
// user costs the same time as a wrong password.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("task-ui"), bcrypt.DefaultCost)

// HashPassword returns the bcrypt hash of a password, as task-ui.yml
// stores it.
func HashPassword(password string) (string, error) {
	out, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(out), err
}

// CheckPassword reports whether a password matches a user. A nil user
// never matches, and takes as long as a real comparison.
func CheckPassword(user *User, password string) bool {
	if user == nil {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil
}
