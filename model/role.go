package model

// Role is the permission set a user holds. The values mirror the CHECK
// constraint in schema/user.up.sql; adding one means adding it in both
// places.
type Role = string

// The roles a user can hold.
const (
	// RoleAdmin may run anything and manage the instance settings.
	RoleAdmin Role = "admin"
	// RoleRunner may run jobs and read history.
	RoleRunner Role = "runner"
	// RoleViewer may read jobs and history, and run nothing.
	RoleViewer Role = "viewer"
)

// Roles returns every role, most permissive first.
func Roles() []Role {
	return []Role{RoleAdmin, RoleRunner, RoleViewer}
}

// ValidRole reports whether role is one of the known roles.
func ValidRole(role Role) bool {
	for _, known := range Roles() {
		if known == role {
			return true
		}
	}
	return false
}
