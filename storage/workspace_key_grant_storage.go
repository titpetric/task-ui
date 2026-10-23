package storage

import (
	"context"
	"errors"
	"slices"

	"github.com/titpetric/task-ui/model"
)

// WorkspaceKeyGrantStorage keeps whose runs in a workspace get one of
// its SSH keys: roles, and users by name. A key with no grant goes to
// every run in the workspace; a grant narrows it to the runs of the
// roles and the users it names. The storage keys a key by its file
// name, which a rotation keeps, so it keeps its grants.
type WorkspaceKeyGrantStorage struct {
	db DB
}

// NewWorkspaceKeyGrantStorage returns the storage over a pool.
func NewWorkspaceKeyGrantStorage(db DB) *WorkspaceKeyGrantStorage {
	return &WorkspaceKeyGrantStorage{db: db}
}

// Set replaces the roles and the users a workspace's key goes to. With
// none of either, every run in the workspace gets it.
func (s *WorkspaceKeyGrantStorage) Set(ctx context.Context, workspaceID, keyName string, roles, users []string) error {
	if workspaceID == "" || keyName == "" {
		return errors.New("a grant needs a workspace and a key")
	}
	for _, role := range roles {
		if !model.ValidRole(role) {
			return errors.New("unknown role " + role)
		}
	}
	if err := s.Delete(ctx, workspaceID, keyName); err != nil {
		return err
	}
	db := client(s.db)
	insert := func(role, user string) error {
		row := &model.WorkspaceKeyGrant{ID: model.NewID(), WorkspaceID: workspaceID, KeyName: keyName, Role: role, UserName: user, CreatedAt: stamp()}
		return db.Insert(ctx, model.WorkspaceKeyGrantTable, row)
	}
	seen := map[string]bool{}
	for _, role := range roles {
		if !seen["role:"+role] {
			seen["role:"+role] = true
			if err := insert(role, ""); err != nil {
				return err
			}
		}
	}
	for _, user := range users {
		if user != "" && !seen["user:"+user] {
			seen["user:"+user] = true
			if err := insert("", user); err != nil {
				return err
			}
		}
	}
	return nil
}

// All returns the grants of every key of a workspace, by key name: the
// roles, most permissive first, and the users, by name.
func (s *WorkspaceKeyGrantStorage) All(ctx context.Context, workspaceID string) (roles, users map[string][]string, err error) {
	rows, err := client(s.db).Select[model.WorkspaceKeyGrant](ctx, "SELECT * FROM workspace_key_grant WHERE workspace_id = ? ORDER BY user_name", workspaceID)
	if err != nil {
		return nil, nil, err
	}
	roles, users = map[string][]string{}, map[string][]string{}
	for _, row := range rows {
		if row.Role != "" {
			roles[row.KeyName] = append(roles[row.KeyName], row.Role)
		} else {
			users[row.KeyName] = append(users[row.KeyName], row.UserName)
		}
	}
	for name, list := range roles {
		roles[name] = byRole(list)
	}
	return roles, users, nil
}

// For returns which of a workspace's keys, in the order given, a run that a user of a role starts gets.
// A key without a grant goes to every run, a granted one to runs of its role or user.
func (s *WorkspaceKeyGrantStorage) For(ctx context.Context, workspaceID string, keyNames []string, role, user string) ([]string, error) {
	roles, users, err := s.All(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, name := range keyNames {
		toRoles, toUsers := roles[name], users[name]
		open := len(toRoles) == 0 && len(toUsers) == 0
		if open || (role != "" && slices.Contains(toRoles, role)) || (user != "" && slices.Contains(toUsers, user)) {
			result = append(result, name)
		}
	}
	return result, nil
}

// Delete forgets a key's grants, when the key goes.
func (s *WorkspaceKeyGrantStorage) Delete(ctx context.Context, workspaceID, keyName string) error {
	return client(s.db).Exec(ctx, "DELETE FROM workspace_key_grant WHERE workspace_id = ? AND key_name = ?", workspaceID, keyName)
}

// DeleteWorkspace forgets every grant of a workspace, when it goes.
func (s *WorkspaceKeyGrantStorage) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	return client(s.db).Exec(ctx, "DELETE FROM workspace_key_grant WHERE workspace_id = ?", workspaceID)
}
