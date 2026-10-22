package api

import (
	"github.com/titpetric/task-ui/permission"
)

// PolicyView is the permission policy of a role or a user: its entries,
// a path to allow or deny, and what the subject may do by them,
// Effective by every path a check names. A user's adds the role and the
// role's entries. Fixed marks the admin role and its users, which may
// do everything and hold no policy to change.
type PolicyView struct {
	Kind       string            `json:"kind"`
	Subject    string            `json:"subject"`
	Role       string            `json:"role,omitempty"`
	Fixed      bool              `json:"fixed"`
	Policy     permission.Policy `json:"policy"`
	RolePolicy permission.Policy `json:"role_policy,omitempty"`
	Effective  map[string]bool   `json:"effective"`
}
