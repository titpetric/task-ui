package api

import (
	"github.com/titpetric/task-ui/config"
)

// PermissionsView is permissions.yml as GET /api/permissions answers it,
// for the permission matrix: the rules every section gets, the sections
// and the application permissions, each with its title, in file order,
// every path a policy sets, and the roles.
type PermissionsView struct {
	Rules       []config.Entry `json:"rules"`
	Sections    []config.Entry `json:"sections"`
	Permissions []config.Entry `json:"permissions"`
	Paths       []string       `json:"paths"`
	Roles       []string       `json:"roles"`
}
