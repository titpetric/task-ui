package view

import (
	"slices"

	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

// UserKeys returns the instance SSH keys each user's runs get, by user name, sorted, as keystore.Local.ForRun
// picks them: linked to the user by name or to the user's effective role; a user with none has no entry.
func UserKeys(keys []keystore.Key, users []model.User) map[string][]string {
	names := map[string][]string{}
	for i := range users {
		user := &users[i]
		role := user.EffectiveRole()
		for _, key := range keys {
			if key.Purpose != keystore.PurposeSSH || key.Workspace() != "" {
				continue
			}
			if slices.Contains(key.Roles, role) || slices.Contains(key.Users, user.Name) {
				names[user.Name] = append(names[user.Name], key.Name)
			}
		}
		slices.Sort(names[user.Name])
	}
	return names
}
