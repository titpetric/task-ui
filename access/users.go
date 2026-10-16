package access

import (
	"fmt"
	"strings"
)

// Users is the configured user list.
type Users []User

// Find returns the user by name, or nil.
func (us Users) Find(name string) *User {
	for i := range us {
		if us[i].Name == name {
			u := us[i]
			return &u
		}
	}
	return nil
}

// Validate checks the list as the configuration loads it: names present
// and unique, roles known, passwords hashed. A missing role is viewer.
func (us Users) Validate() error {
	seen := map[string]bool{}
	for i := range us {
		user := &us[i]
		if user.Name == "" {
			return fmt.Errorf("users[%d]: name is required", i)
		}
		if seen[user.Name] {
			return fmt.Errorf("users[%d]: duplicate user %q", i, user.Name)
		}
		seen[user.Name] = true
		if user.Role == "" {
			user.Role = RoleViewer
		} else if !user.Role.Valid() {
			return fmt.Errorf("users[%d]: unknown role %q", i, user.Role)
		}
		if !strings.HasPrefix(user.Password, "$2") {
			return fmt.Errorf("users[%d]: password must be a bcrypt hash, generate one with `task-ui hash`", i)
		}
	}
	return nil
}
