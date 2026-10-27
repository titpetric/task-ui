package view

import (
	"slices"

	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// RoleMembers is what a role holds: the users who act with it and the
// instance SSH keys linked to it.
type RoleMembers struct {
	Users []string
	Keys  []string
}

// NewRoleMembers returns the members of each role, by role name, every
// role with an entry. A user counts under their effective role, so an
// admin flag puts them under admin; a workspace's key and an encryption
// key belong to no role. Names are sorted.
func NewRoleMembers(users []model.User, keys []keystore.Key) map[string]RoleMembers {
	members := map[string]RoleMembers{}
	for _, role := range model.Roles() {
		members[role] = RoleMembers{}
	}
	for i := range users {
		role := users[i].EffectiveRole()
		m := members[role]
		m.Users = append(m.Users, users[i].Name)
		members[role] = m
	}
	for _, key := range keys {
		if key.Purpose != keystore.PurposeSSH || key.Workspace() != "" {
			continue
		}
		for _, role := range key.Roles {
			m := members[role]
			m.Keys = append(m.Keys, key.Name)
			members[role] = m
		}
	}
	for role, m := range members {
		slices.Sort(m.Users)
		slices.Sort(m.Keys)
		members[role] = m
	}
	return members
}
