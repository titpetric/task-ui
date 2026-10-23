package app

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// UsersFile is the password file, htpasswd style: who may sign in. The
// file is the truth for who: a user it does not name cannot sign in and
// leaves storage, and a user it names gets the password it holds.
// Storage holds what they are, the admin flag and the role: the first
// read of a file into empty storage makes every user an admin, a file
// of one user makes that user an admin while storage holds none, and a
// name new to a storage with users comes in as a viewer. Watch reads it
// again when it changes.
type UsersFile struct {
	Path string
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
// file, as for any new user of the file.
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

// Remove takes a user's line out of the file and the user out of
// storage at once, with their sessions, tokens and policy, so no request
// of theirs passes after it. A name the file does not hold is
// storage.ErrNotFound.
func (f *UsersFile) Remove(ctx context.Context, users *storage.UserStorage, passwords *storage.UserAuthPasswordStorage, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	removed, err := access.RemoveHtpasswd(f.Path, name)
	if err != nil {
		return err
	}
	user, err := users.GetByName(ctx, name)
	if err != nil {
		return err
	}
	if !removed {
		return storage.ErrNotFound
	}
	if err := users.Delete(ctx, user.ID); err != nil {
		return err
	}
	_, _, err = f.sync(ctx, users, passwords)
	return err
}

// NewUsersFile returns the file.
func NewUsersFile(path string) *UsersFile {
	return &UsersFile{Path: path, Interval: 5 * time.Second}
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

// Sync writes the file to storage, as UsersFile says, and returns how many users it created and removed.
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

// keep creates the users the file names, and updates the passwords
// that changed. Into empty storage every user comes as an admin; into
// storage with users a new one comes as a viewer, the lowest role; and
// a file of one user makes that user an admin when no user is one. It
// returns the count of new users.
func (f *UsersFile) keep(ctx context.Context, users *storage.UserStorage, passwords *storage.UserAuthPasswordStorage, creds []access.Credential, existing []model.User) (int, error) {
	byName := map[string]*model.User{}
	admins := 0
	for i := range existing {
		byName[existing[i].Name] = &existing[i]
		if existing[i].Admin {
			admins++
		}
	}
	created := 0
	for _, cred := range creds {
		user, ok := byName[cred.Name]
		if !ok {
			user = &model.User{Name: cred.Name, Role: model.RoleViewer, Admin: len(existing) == 0}
			if err := users.Create(ctx, user); err != nil {
				return created, err
			}
			if user.Admin {
				slog.Info("first user of the password file made an admin", "user", user.Name)
			}
			created++
		}
		if len(creds) == 1 && admins == 0 && !user.Admin {
			// The row again, not the listing: the folders may have
			// changed under Settings since.
			current, err := users.GetByName(ctx, cred.Name)
			if err != nil {
				return created, err
			}
			current.Admin = true
			if err := users.Update(ctx, current); err != nil {
				return created, err
			}
			slog.Info("the one user of the password file made an admin, as no user is one", "user", user.Name)
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
