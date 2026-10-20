package storage

import (
	"context"
	"errors"
	"slices"

	"github.com/titpetric/task-ui/model"
)

// SSHKeyPolicyStorage keeps whose runs get an instance SSH key: roles,
// and users by name, the way allow lists name them. The storage keys a key
// by its file name, which a rotation keeps, so it keeps its policies. A
// key can be in many roles' and users' sets.
type SSHKeyPolicyStorage struct {
	db DB
}

// NewSSHKeyPolicyStorage returns the storage over a pool.
func NewSSHKeyPolicyStorage(db DB) *SSHKeyPolicyStorage {
	return &SSHKeyPolicyStorage{db: db}
}

// Set replaces the roles that get a key. No roles leaves it with none.
func (s *SSHKeyPolicyStorage) Set(ctx context.Context, keyName string, roles []string) error {
	if keyName == "" {
		return errors.New("a policy needs a key")
	}
	for _, role := range roles {
		if !model.ValidRole(role) {
			return errors.New("unknown role " + role)
		}
	}
	db := client(s.db)
	if err := db.Exec(ctx, "DELETE FROM ssh_key_policy WHERE key_name = ?", keyName); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, role := range roles {
		if seen[role] {
			continue
		}
		seen[role] = true
		row := &model.SSHKeyPolicy{ID: model.NewID(), KeyName: keyName, Role: role, CreatedAt: stamp()}
		if err := db.Insert(ctx, model.SSHKeyPolicyTable, row); err != nil {
			return err
		}
	}
	return nil
}

// Roles returns the roles that get a key, most permissive first.
func (s *SSHKeyPolicyStorage) Roles(ctx context.Context, keyName string) ([]string, error) {
	all, err := s.All(ctx)
	if err != nil {
		return nil, err
	}
	return all[keyName], nil
}

// All returns every key's roles, by key name, most permissive first.
func (s *SSHKeyPolicyStorage) All(ctx context.Context) (map[string][]string, error) {
	rows, err := client(s.db).Select[model.SSHKeyPolicy](ctx, "SELECT * FROM ssh_key_policy")
	if err != nil {
		return nil, err
	}
	result := map[string][]string{}
	for _, row := range rows {
		result[row.KeyName] = append(result[row.KeyName], row.Role)
	}
	for name, roles := range result {
		result[name] = byRole(roles)
	}
	return result, nil
}

// Keys returns the names of the keys a role gets, by name.
func (s *SSHKeyPolicyStorage) Keys(ctx context.Context, role string) ([]string, error) {
	return client(s.db).Select[string](ctx, "SELECT key_name FROM ssh_key_policy WHERE role = ? ORDER BY key_name", role)
}

// SetUsers replaces the users that get a key by name. No users leaves it
// with none by name; its roles stay.
func (s *SSHKeyPolicyStorage) SetUsers(ctx context.Context, keyName string, users []string) error {
	if keyName == "" {
		return errors.New("a policy needs a key")
	}
	db := client(s.db)
	if err := db.Exec(ctx, "DELETE FROM ssh_key_user WHERE key_name = ?", keyName); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, user := range users {
		if user == "" || seen[user] {
			continue
		}
		seen[user] = true
		row := &model.SSHKeyUser{ID: model.NewID(), KeyName: keyName, UserName: user, CreatedAt: stamp()}
		if err := db.Insert(ctx, model.SSHKeyUserTable, row); err != nil {
			return err
		}
	}
	return nil
}

// AllUsers returns every key's users, by key name, each by name.
func (s *SSHKeyPolicyStorage) AllUsers(ctx context.Context) (map[string][]string, error) {
	rows, err := client(s.db).Select[model.SSHKeyUser](ctx, "SELECT * FROM ssh_key_user ORDER BY user_name")
	if err != nil {
		return nil, err
	}
	result := map[string][]string{}
	for _, row := range rows {
		result[row.KeyName] = append(result[row.KeyName], row.UserName)
	}
	return result, nil
}

// For returns the names of the keys a user's runs get: those given to
// their role and those given to them, each once, by name.
func (s *SSHKeyPolicyStorage) For(ctx context.Context, role, user string) ([]string, error) {
	return client(s.db).Select[string](ctx,
		"SELECT key_name FROM ssh_key_policy WHERE role = ? UNION SELECT key_name FROM ssh_key_user WHERE user_name = ? ORDER BY key_name", role, user)
}

// Delete forgets a key's roles and users, when the key goes.
func (s *SSHKeyPolicyStorage) Delete(ctx context.Context, keyName string) error {
	db := client(s.db)
	if err := db.Exec(ctx, "DELETE FROM ssh_key_user WHERE key_name = ?", keyName); err != nil {
		return err
	}
	return db.Exec(ctx, "DELETE FROM ssh_key_policy WHERE key_name = ?", keyName)
}

// byRole orders roles as model.Roles does, most permissive first.
func byRole(roles []string) []string {
	ordered := []string{}
	for _, role := range model.Roles() {
		if slices.Contains(roles, role) {
			ordered = append(ordered, role)
		}
	}
	return ordered
}
