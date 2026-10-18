// Package access is who may do what: the passwords users sign in with
// and the sessions they carry. The users themselves are rows in storage,
// and the rules on them are on model.User.
package access

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/titpetric/task-ui/model"
)

// dummyHash is compared against when a login names no user, so a missing
// user costs the same time as a wrong password.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("task-ui"), bcrypt.DefaultCost)

// HashPassword returns the bcrypt hash of a password, as the user row
// stores it.
func HashPassword(password string) (string, error) {
	out, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(out), err
}

// CheckPassword reports whether a password matches a user. A nil user
// never matches, and takes as long as a real comparison.
func CheckPassword(user *model.User, password string) bool {
	if user == nil {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil
}

// Hashed reports whether text is a bcrypt hash, as a configured password
// must be.
func Hashed(text string) bool {
	return len(text) > 4 && text[0] == '$' && text[1] == '2'
}
