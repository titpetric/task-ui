package view

import (
	"slices"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
)

// KeyHolder is a row of the keys page's policy table: a role, or a user,
// and the instance keys its runs get. A user's runs get the keys given to
// them by name, Direct, and those of their role, ViaRole.
type KeyHolder struct {
	Role    string
	User    *model.User
	Direct  []sshkey.Key
	ViaRole []sshkey.Key
}

// KeyHolders lays out who gets which key: a row for each role, then one
// for each user. byRole and byUser name, for each key, the roles and the
// users that get it; backups go to no one.
func KeyHolders(keys []sshkey.Key, byRole, byUser map[string][]string, users []model.User) []KeyHolder {
	var rows []KeyHolder
	forRole := func(role string) []sshkey.Key {
		var list []sshkey.Key
		for _, key := range keys {
			if !key.Backup && slices.Contains(byRole[key.Name], role) {
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
			if !key.Backup && slices.Contains(byUser[key.Name], user.Name) {
				row.Direct = append(row.Direct, key)
			}
		}
		rows = append(rows, row)
	}
	return rows
}
