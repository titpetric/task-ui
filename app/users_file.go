package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/titpetric/task-ui/storage"
)

// UsersFile is the shape of task-ui.yml: the users to seed. The file is
// the way in on a fresh instance; the users page takes over after.
type UsersFile struct {
	Users []UserEntry `yaml:"users"`
}

// LoadUsersFile reads and validates a users file.
func LoadUsersFile(path string) (*UsersFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file UsersFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &file, file.Validate()
}

// Validate checks every entry and that the names are unique.
func (f *UsersFile) Validate() error {
	seen := map[string]bool{}
	for _, entry := range f.Users {
		if err := entry.Validate(); err != nil {
			return err
		}
		if seen[entry.Name] {
			return errors.New("user " + entry.Name + " is listed twice")
		}
		seen[entry.Name] = true
	}
	return nil
}

// Seed writes the entries to storage: a user that is not there is
// created, one that is gets the role, folders and password of the file.
// It returns how many were created.
func (f *UsersFile) Seed(ctx context.Context, users *storage.UserStorage) (int, error) {
	created := 0
	for _, entry := range f.Users {
		user, err := entry.User()
		if err != nil {
			return created, err
		}
		existing, err := users.GetByName(ctx, user.Name)
		if errors.Is(err, storage.ErrNotFound) {
			if err := users.Create(ctx, user); err != nil {
				return created, err
			}
			created++
			continue
		}
		if err != nil {
			return created, err
		}
		existing.Role, existing.Folders, existing.Password = user.Role, user.Folders, user.Password
		if err := users.Update(ctx, existing); err != nil {
			return created, err
		}
	}
	return created, nil
}
