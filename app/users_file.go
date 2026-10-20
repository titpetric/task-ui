package app

import (
	"context"
	"errors"
	"os"
	"regexp"
	"slices"
	"sync"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// UsersFile is the password file, htpasswd style, and the role lists
// from the environment: who may sign in, and what they are. The file is
// the truth: a user it does not name cannot sign in and leaves
// storage; a user it names gets the password it holds and the role the
// lists set. Watch reads it again when it changes.
type UsersFile struct {
	Path string
	// Admins and Viewers name the users with those roles; everyone else
	// in the file is a runner.
	Admins  []string
	Viewers []string
	// Interval is how often Watch looks at the file. Zero is five seconds.
	Interval time.Duration

	mu       sync.Mutex
	modified time.Time
}

// ErrUserExists reports an Add of a name the file holds.
var ErrUserExists = api.ErrUserExists

// userName is what a new user's name may hold: it names their keys too,
// id_<type>_<name>.
var userName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)

// Add puts a new user in the file, with the password hashed, and reads
// the file again so they can sign in at once. The role comes from the
// lists, as for any user of the file.
func (f *UsersFile) Add(ctx context.Context, users *storage.UserStorage, passwords *storage.UserAuthPasswordStorage, name, password string) error {
	if !userName.MatchString(name) {
		return errors.New("a user name is lowercase letters, digits and dashes, up to 32")
	}
	if len(password) < 8 {
		return errors.New("a password needs at least 8 characters")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	creds, err := access.LoadHtpasswd(f.Path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, cred := range creds {
		if cred.Name == name {
			return ErrUserExists
		}
	}
	hash, err := access.HashPassword(password)
	if err != nil {
		return err
	}
	if err := access.SetHtpasswd(f.Path, name, hash); err != nil {
		return err
	}
	_, _, err = f.sync(ctx, users, passwords)
	return err
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
	f.mu.Lock()
	defer f.mu.Unlock()
	return !info.ModTime().Equal(f.modified)
}

// Sync writes the file to storage: it creates or updates the users it
// names with their role, with their hash as their password, and removes
// the users it does not name, with what they sign in with. It returns
// how many users it created and removed.
func (f *UsersFile) Sync(ctx context.Context, users *storage.UserStorage, passwords *storage.UserAuthPasswordStorage) (int, int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.sync(ctx, users, passwords)
}

// sync is Sync, with the lock held.
func (f *UsersFile) sync(ctx context.Context, users *storage.UserStorage, passwords *storage.UserAuthPasswordStorage) (int, int, error) {
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
	created, err := f.keep(ctx, users, passwords, creds, existing)
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

// keep creates the users the file names, and updates the roles and
// passwords that changed. It returns the count of new users.
func (f *UsersFile) keep(ctx context.Context, users *storage.UserStorage, passwords *storage.UserAuthPasswordStorage, creds []access.Credential, existing []model.User) (int, error) {
	byName := map[string]*model.User{}
	for i := range existing {
		byName[existing[i].Name] = &existing[i]
	}
	created := 0
	for _, cred := range creds {
		role := f.Role(cred.Name)
		user, ok := byName[cred.Name]
		if !ok {
			user = &model.User{Name: cred.Name, Role: role}
			if err := users.Create(ctx, user); err != nil {
				return created, err
			}
			created++
		} else if user.Role != role {
			// The row again, not the listing: the folders may have
			// changed under Settings since.
			current, err := users.GetByName(ctx, cred.Name)
			if err != nil {
				return created, err
			}
			current.Role = role
			if err := users.Update(ctx, current); err != nil {
				return created, err
			}
		}
		hash, err := passwords.PasswordHash(ctx, user.ID)
		if err != nil {
			return created, err
		}
		if hash != cred.Hash {
			if err := passwords.SetPassword(ctx, user.ID, cred.Hash); err != nil {
				return created, err
			}
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
func (f *UsersFile) Watch(ctx context.Context, users *storage.UserStorage, passwords *storage.UserAuthPasswordStorage, report func(error)) {
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
		if _, _, err := f.Sync(ctx, users, passwords); err != nil && !errors.Is(err, context.Canceled) {
			report(err)
		}
	}
}
