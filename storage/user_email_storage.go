package storage

import (
	"context"
	"errors"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// UserEmailStorage keeps the email addresses users sign in with. A user
// may have several; an address belongs to one user.
type UserEmailStorage struct {
	db DB
}

// NewUserEmailStorage returns the storage over a pool.
func NewUserEmailStorage(db DB) *UserEmailStorage {
	return &UserEmailStorage{db: db}
}

// Add gives a user an email address, kept lower case.
func (s *UserEmailStorage) Add(ctx context.Context, userID, email string) error {
	email = normalEmail(email)
	if userID == "" {
		return errors.New("an email address needs a user")
	}
	if at := strings.IndexByte(email, '@'); at < 1 || at == len(email)-1 || strings.ContainsAny(email, " \t\n") {
		return errors.New("not an email address: " + email)
	}
	row := &model.UserEmail{ID: model.NewID(), UserID: userID, Email: email, CreatedAt: stamp()}
	row.UpdatedAt = row.CreatedAt
	return client(s.db).Insert(ctx, model.UserEmailTable, row)
}

// List returns a user's email addresses, oldest first.
func (s *UserEmailStorage) List(ctx context.Context, userID string) ([]model.UserEmail, error) {
	return client(s.db).Select[model.UserEmail](ctx, "SELECT * FROM user_email WHERE user_id = ? ORDER BY rowid", userID)
}

// UserID returns the user an email address belongs to.
func (s *UserEmailStorage) UserID(ctx context.Context, email string) (string, error) {
	id, err := client(s.db).Get[string](ctx, "SELECT user_id FROM user_email WHERE email = ?", normalEmail(email))
	if err != nil {
		return "", notFound(err)
	}
	return *id, nil
}

// Delete removes an email address of a user.
func (s *UserEmailStorage) Delete(ctx context.Context, userID, email string) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_email WHERE user_id = ? AND email = ?", userID, normalEmail(email))
}

// DeleteByUser removes every email address of a user.
func (s *UserEmailStorage) DeleteByUser(ctx context.Context, userID string) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_email WHERE user_id = ?", userID)
}

func normalEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
