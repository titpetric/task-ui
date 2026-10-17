package app

import (
	"context"
	"errors"
	"os"
	"slices"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// UsersFile is the password file, htpasswd style, and the role lists
// from the environment: who may sign in, and what they are. The file is
// the truth: a user it does not name cannot sign in and is removed from
// storage; a user it names gets the password it holds and the role the
// lists give. It is re-read when it changes.
type UsersFile struct {
	Path string
	// Admins and Viewers name the users with those roles; everyone else
	// in the file is a runner.
	Admins  []string
	Viewers []string
	// Interval is how often Watch looks at the file. Zero is five seconds.
	Interval time.Duration

	modified time.Time
}

// NewUsersFile returns the file with the role lists.
func NewUsersFile(path string, admins, viewers []string) *UsersFile {
	return &UsersFile{Path: path, Admins: admins, Viewers: viewers, Interval: 5 * time.Second}
}

// Role returns the role of a user by name: the lists decide, admins first.
func (f *UsersFile) Role(name string) string {
	switch {
	case slices.Contains(f.Admins, name):
		return model.RoleAdmin
	case slices.Contains(f.Viewers, name):
		return model.RoleViewer
	}
	return model.RoleRunner
}

// Changed reports whether the file changed since the last Sync.
func (f *UsersFile) Changed() bool {
	info, err := os.Stat(f.Path)
	if err != nil {
		return false
	}
	return !info.ModTime().Equal(f.modified)
}

// Sync writes the file to storage: users it names are created or
// updated with its hash and their role, users it does not name are
// removed. It returns how many were created and removed.
func (f *UsersFile) Sync(ctx context.Context, users *storage.UserStorage) (int, int, error) {
	info, err := os.Stat(f.Path)
	if err != nil {
		return 0, 0, err
	}
	creds, err := access.LoadHtpasswd(f.Path)
	if err != nil {
		return 0, 0, err
	}
	existing, err := users.List(ctx)
	if err != nil {
		return 0, 0, err
	}
	created, err := f.keep(ctx, users, creds, existing)
	if err != nil {
		return created, 0, err
	}
	removed, err := drop(ctx, users, creds, existing)
	if err != nil {
		return created, removed, err
	}
	f.modified = info.ModTime()
	return created, removed, nil
}

// keep creates the users the file names and updates the ones that
// changed. It returns how many were created.
func (f *UsersFile) keep(ctx context.Context, users *storage.UserStorage, creds []access.Credential, existing []model.User) (int, error) {
	byName := map[string]*model.User{}
	for i := range existing {
		byName[existing[i].Name] = &existing[i]
	}
	created := 0
	for _, cred := range creds {
		role := f.Role(cred.Name)
		user, ok := byName[cred.Name]
		if !ok {
			user = &model.User{Name: cred.Name, Password: cred.Hash, Role: role}
			if err := users.Create(ctx, user); err != nil {
				return created, err
			}
			created++
			continue
		}
		if user.Password == cred.Hash && user.Role == role {
			continue
		}
		user.Password, user.Role = cred.Hash, role
		if err := users.Update(ctx, user); err != nil {
			return created, err
		}
	}
	return created, nil
}

// drop removes the users the file does not name. It returns how many.
func drop(ctx context.Context, users *storage.UserStorage, creds []access.Credential, existing []model.User) (int, error) {
	named := map[string]bool{}
	for _, cred := range creds {
		named[cred.Name] = true
	}
	removed := 0
	for _, user := range existing {
		if named[user.Name] {
			continue
		}
		if err := users.Delete(ctx, user.ID); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

// Watch syncs the file whenever it changes, until the context ends. A
// file that stops parsing keeps the users as they are: a typo must not
// lock everyone out.
func (f *UsersFile) Watch(ctx context.Context, users *storage.UserStorage, report func(error)) {
	interval := f.Interval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if !f.Changed() {
			continue
		}
		if _, _, err := f.Sync(ctx, users); err != nil && !errors.Is(err, context.Canceled) {
			report(err)
		}
	}
}
