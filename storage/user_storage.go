package storage

import (
	"context"
	"errors"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// UserStorage keeps the users, and the permission policies of the users
// and their roles.
type UserStorage struct {
	db       DB
	policies *PermissionPolicyStorage
}

// NewUserStorage returns the storage over a pool.
func NewUserStorage(db DB) *UserStorage {
	return &UserStorage{db: db, policies: NewPermissionPolicyStorage(db)}
}

// Policies returns the permission policies, the one cache of them.
func (s *UserStorage) Policies() *PermissionPolicyStorage {
	return s.policies
}

// Permissions returns what a user may do.
func (s *UserStorage) Permissions(ctx context.Context, user *model.User) permission.Set {
	return s.policies.For(ctx, user)
}

// Create inserts a user. It fills in an empty ID and sets the timestamps.
func (s *UserStorage) Create(ctx context.Context, user *model.User) error {
	if user.Name == "" {
		return errors.New("a user needs a name")
	}
	if !model.ValidRole(user.Role) {
		return errors.New("unknown role " + user.Role)
	}
	if user.ID == "" {
		user.ID = model.NewID()
	}
	if user.Folders == "" {
		user.Folders = model.EncodeList(nil)
	}
	user.CreatedAt = stamp()
	user.UpdatedAt = user.CreatedAt
	return client(s.db).Insert(ctx, model.UserTable, user)
}

// Get returns a user by ID.
func (s *UserStorage) Get(ctx context.Context, id string) (*model.User, error) {
	user, err := client(s.db).Get[model.User](ctx, "SELECT * FROM user WHERE id = ?", id)
	return user, notFound(err)
}

// GetByName returns a user by name.
func (s *UserStorage) GetByName(ctx context.Context, name string) (*model.User, error) {
	user, err := client(s.db).Get[model.User](ctx, "SELECT * FROM user WHERE name = ?", name)
	return user, notFound(err)
}

// List returns every user, by name.
func (s *UserStorage) List(ctx context.Context) ([]model.User, error) {
	return client(s.db).Select[model.User](ctx, "SELECT * FROM user ORDER BY name")
}

// Count returns how many users there are.
func (s *UserStorage) Count(ctx context.Context) (int64, error) {
	count, err := client(s.db).Get[int64](ctx, "SELECT COUNT(*) FROM user")
	if err != nil {
		return 0, err
	}
	return *count, nil
}

// Update writes a user back, by ID.
func (s *UserStorage) Update(ctx context.Context, user *model.User) error {
	if !model.ValidRole(user.Role) {
		return errors.New("unknown role " + user.Role)
	}
	user.UpdatedAt = stamp()
	return client(s.db).Update(ctx, model.UserTable, user, "id")
}

// Delete removes a user by ID, with the email addresses, passwords and
// tokens it signs in with, the sessions it has open and its permission
// policy. Runs, notes and history keep its name as text.
func (s *UserStorage) Delete(ctx context.Context, id string) error {
	if err := NewUserEmailStorage(s.db).DeleteByUser(ctx, id); err != nil {
		return err
	}
	if err := NewUserSessionStorage(s.db).DeleteByUser(ctx, id); err != nil {
		return err
	}
	if err := NewUserAuthPasswordStorage(s.db).DeleteByUser(ctx, id); err != nil {
		return err
	}
	if err := NewUserAuthTokenStorage(s.db).DeleteByUser(ctx, id); err != nil {
		return err
	}
	if user, err := s.Get(ctx, id); err == nil {
		if err := s.policies.DeleteSubject(ctx, permission.KindUser, user.Name); err != nil {
			return err
		}
	}
	return client(s.db).Exec(ctx, "DELETE FROM user WHERE id = ?", id)
}

// MoveFolder rewrites the folder lists of the users that name a folder,
// or one inside it, to name it under its new ID: access follows a
// project that moves. It returns how many users changed.
func (s *UserStorage) MoveFolder(ctx context.Context, from, to string) (int, error) {
	users, err := s.List(ctx)
	if err != nil {
		return 0, err
	}
	changed := 0
	for i := range users {
		list, moved := users[i].FolderList(), false
		for j, folder := range list {
			if id, ok := model.Moved(folder, from, to); ok {
				list[j], moved = id, true
			}
		}
		if !moved {
			continue
		}
		users[i].Folders = model.EncodeList(list)
		if err := s.Update(ctx, &users[i]); err != nil {
			return changed, err
		}
		changed++
	}
	return changed, nil
}
