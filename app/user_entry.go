package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
)

// UserEntry is one user of the users file. The password is a bcrypt hash,
// as task-ui hash prints it, or a plain password that is hashed when the
// user is seeded.
type UserEntry struct {
	Name     string   `yaml:"name"`
	Password string   `yaml:"password"`
	Role     string   `yaml:"role"`
	Folders  []string `yaml:"folders"`
}

// Validate checks the entry.
func (e UserEntry) Validate() error {
	if strings.TrimSpace(e.Name) == "" {
		return errors.New("a user needs a name")
	}
	if e.Password == "" {
		return fmt.Errorf("user %s: a password is needed", e.Name)
	}
	if e.Role != "" && !model.ValidRole(e.Role) {
		return fmt.Errorf("user %s: role %q is not one of %s", e.Name, e.Role, strings.Join(model.Roles(), ", "))
	}
	return nil
}

// Hash returns the password as a hash, hashing a plain one.
func (e UserEntry) Hash() (string, error) {
	if access.Hashed(e.Password) {
		return e.Password, nil
	}
	return access.HashPassword(e.Password)
}

// User returns the entry as a user row, with the defaults filled in.
func (e UserEntry) User() (*model.User, error) {
	hash, err := e.Hash()
	if err != nil {
		return nil, err
	}
	role := e.Role
	if role == "" {
		role = model.RoleViewer
	}
	return &model.User{Name: strings.TrimSpace(e.Name), Password: hash, Role: role, Folders: model.EncodeList(e.Folders)}, nil
}
