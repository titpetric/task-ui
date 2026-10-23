package view

import (
	"slices"

	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// KeyHolder is a row of the keys page's policy table: a role, or a user,
// and the instance SSH keys its runs get. A user's runs get the keys given to
// them by name, Direct, and those of their role, ViaRole.
type KeyHolder struct {
	Role    string
	User    *model.User
	Direct  []keystore.Key
	ViaRole []keystore.Key
}

// KeyHolders lays out who gets which instance SSH key, a row for each role and then for each user.
func KeyHolders(keys []keystore.Key, users []model.User) []KeyHolder {
	var rows []KeyHolder
	forRole := func(role string) []keystore.Key {
		var list []keystore.Key
		for _, key := range keys {
			if key.Purpose == keystore.PurposeSSH && key.Workspace() == "" && slices.Contains(key.Roles, role) {
				list = append(list, key)
			}
		}
		return list
	}
	for _, role := range model.Roles() {
		rows = append(rows, KeyHolder{Role: role, Direct: forRole(role)})
	}
	for i := range users {
		user := &users[i]
		row := KeyHolder{Role: user.EffectiveRole(), User: user, ViaRole: forRole(user.EffectiveRole())}
		for _, key := range keys {
			if key.Purpose == keystore.PurposeSSH && key.Workspace() == "" && slices.Contains(key.Users, user.Name) {
				row.Direct = append(row.Direct, key)
			}
		}
		rows = append(rows, row)
	}
	return rows
}
