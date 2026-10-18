package access

import (
	"context"
	"errors"
	"strings"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// ErrWrongCredentials is the one answer to a failed sign in, whatever
// was wrong: the login, the password, or a backup password already used.
var ErrWrongCredentials = errors.New("wrong username or password")

// Authenticator checks a sign in: a login, which is a user name or one
// of the user's email addresses, and a password, which is the user's
// password or one of the unused backup passwords. A backup password
// that signs in is used up.
type Authenticator struct {
	users     *storage.UserStorage
	emails    *storage.UserEmailStorage
	passwords *storage.UserAuthPasswordStorage
}

// NewAuthenticator returns the authenticator over the users and the
// credentials they sign in with.
func NewAuthenticator(users *storage.UserStorage, emails *storage.UserEmailStorage, passwords *storage.UserAuthPasswordStorage) *Authenticator {
	return &Authenticator{users: users, emails: emails, passwords: passwords}
}

// Authenticate returns the user a login and password sign in as.
func (a *Authenticator) Authenticate(ctx context.Context, login, password string) (*model.User, error) {
	user, err := a.lookup(ctx, strings.TrimSpace(login))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			CheckHash("", password)
			return nil, ErrWrongCredentials
		}
		return nil, err
	}
	stored, err := a.passwords.List(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	for _, p := range stored {
		if !p.Usable() || !CheckHash(p.Hash, password) {
			continue
		}
		if p.Kind == model.AuthBackup {
			ok, err := a.passwords.Use(ctx, p.ID)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
		}
		return user, nil
	}
	if len(stored) == 0 {
		CheckHash("", password)
	}
	return nil, ErrWrongCredentials
}

// lookup finds the user by name, then by email address.
func (a *Authenticator) lookup(ctx context.Context, login string) (*model.User, error) {
	if login == "" {
		return nil, storage.ErrNotFound
	}
	user, err := a.users.GetByName(ctx, login)
	if !errors.Is(err, storage.ErrNotFound) || !strings.Contains(login, "@") {
		return user, err
	}
	id, err := a.emails.UserID(ctx, login)
	if err != nil {
		return nil, err
	}
	return a.users.Get(ctx, id)
}
