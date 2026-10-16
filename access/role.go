// Package access decides who may see and run what: users and their roles,
// the passwords they sign in with, and the sessions they carry.
package access

// Role names the permission sets a user can hold.
type Role string

// The roles a user can hold.
const (
	// RoleAdmin may run anything and manage the instance settings.
	RoleAdmin Role = "admin"
	// RoleRunner may run jobs and read history.
	RoleRunner Role = "runner"
	// RoleViewer may read jobs and history, and run nothing.
	RoleViewer Role = "viewer"
)

// Valid reports whether the role is one of the known roles.
func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleRunner, RoleViewer:
		return true
	}
	return false
}
