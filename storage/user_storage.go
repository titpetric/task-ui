package storage

import (
	"context"
	"errors"

	"github.com/titpetric/task-ui/model"
)

// UserStorage keeps the users.
type UserStorage struct {
	db DB
}

// NewUserStorage returns the storage over a pool.
func NewUserStorage(db DB) *UserStorage {
	return &UserStorage{db: db}
}

// Create inserts a user. An empty ID is filled in; the timestamps are set.
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

// Delete removes a user by ID.
func (s *UserStorage) Delete(ctx context.Context, id string) error {
	return client(s.db).Exec(ctx, "DELETE FROM user WHERE id = ?", id)
}
