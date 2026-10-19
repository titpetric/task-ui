package storage

import (
	"context"
	"errors"
	"slices"

	"github.com/titpetric/task-ui/model"
)

// SSHKeyPolicyStorage keeps which roles' runs get an instance SSH key,
// by the key's file name: a rotation keeps the name, so it keeps its
// roles. A key can be in several roles' sets.
type SSHKeyPolicyStorage struct {
	db DB
}

// NewSSHKeyPolicyStorage returns the storage over a pool.
func NewSSHKeyPolicyStorage(db DB) *SSHKeyPolicyStorage {
	return &SSHKeyPolicyStorage{db: db}
}

// Set replaces the roles a key is given to. No roles gives it to none.
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

// Roles returns the roles a key is given to, most permissive first.
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

// Keys returns the names of the keys a role is given, by name.
func (s *SSHKeyPolicyStorage) Keys(ctx context.Context, role string) ([]string, error) {
	return client(s.db).Select[string](ctx, "SELECT key_name FROM ssh_key_policy WHERE role = ? ORDER BY key_name", role)
}

// Delete forgets a key's roles, when the key goes.
func (s *SSHKeyPolicyStorage) Delete(ctx context.Context, keyName string) error {
	return client(s.db).Exec(ctx, "DELETE FROM ssh_key_policy WHERE key_name = ?", keyName)
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
