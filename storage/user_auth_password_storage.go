package storage

import (
	"context"
	"errors"

	"github.com/titpetric/task-ui/model"
)

// UserAuthPasswordStorage keeps the password hashes users sign in with,
// apart from the user rows, so nothing that reads a user reads a hash.
type UserAuthPasswordStorage struct {
	db DB
}

// NewUserAuthPasswordStorage returns the storage over a pool.
func NewUserAuthPasswordStorage(db DB) *UserAuthPasswordStorage {
	return &UserAuthPasswordStorage{db: db}
}

// List returns a user's passwords: the password first, then the backup
// passwords, oldest first.
func (s *UserAuthPasswordStorage) List(ctx context.Context, userID string) ([]model.UserAuthPassword, error) {
	return client(s.db).Select[model.UserAuthPassword](ctx,
		"SELECT * FROM user_auth_password WHERE user_id = ? ORDER BY kind = 'backup', id", userID)
}

// PasswordHash returns the hash of a user's password, empty when there is none.
func (s *UserAuthPasswordStorage) PasswordHash(ctx context.Context, userID string) (string, error) {
	hash, err := client(s.db).Get[string](ctx,
		"SELECT hash FROM user_auth_password WHERE user_id = ? AND kind = 'password' ORDER BY id DESC LIMIT 1", userID)
	if err := notFound(err); errors.Is(err, ErrNotFound) {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return *hash, nil
}

// SetPassword replaces a user's password with a hash.
func (s *UserAuthPasswordStorage) SetPassword(ctx context.Context, userID, hash string) error {
	return s.replace(ctx, userID, model.AuthPassword, []string{hash})
}

// SetBackups replaces a user's backup passwords with hashes, used or not.
func (s *UserAuthPasswordStorage) SetBackups(ctx context.Context, userID string, hashes []string) error {
	return s.replace(ctx, userID, model.AuthBackup, hashes)
}

// Use marks a backup password used. It reports false when it already
// was, so of two sign ins racing on one backup password only one wins.
func (s *UserAuthPasswordStorage) Use(ctx context.Context, id string) (bool, error) {
	db := client(s.db)
	now := stamp()
	if err := db.Exec(ctx, "UPDATE user_auth_password SET used_at = ?, updated_at = ? WHERE id = ? AND kind = 'backup' AND used_at IS NULL", now, now, id); err != nil {
		return false, err
	}
	return db.RowsAffected() == 1, nil
}

// DeleteByUser removes every password of a user.
func (s *UserAuthPasswordStorage) DeleteByUser(ctx context.Context, userID string) error {
	return client(s.db).Exec(ctx, "DELETE FROM user_auth_password WHERE user_id = ?", userID)
}

// replace swaps the user's passwords of one kind for new hashes.
func (s *UserAuthPasswordStorage) replace(ctx context.Context, userID string, kind model.AuthKind, hashes []string) error {
	if userID == "" {
		return errors.New("a password needs a user")
	}
	for _, hash := range hashes {
		if hash == "" {
			return errors.New("a password needs a hash")
		}
	}
	db := client(s.db)
	if err := db.Exec(ctx, "DELETE FROM user_auth_password WHERE user_id = ? AND kind = ?", userID, kind); err != nil {
		return err
	}
	for _, hash := range hashes {
		row := &model.UserAuthPassword{ID: model.NewID(), UserID: userID, Hash: hash, Kind: kind, CreatedAt: stamp()}
		row.UpdatedAt = row.CreatedAt
		if err := db.Insert(ctx, model.UserAuthPasswordTable, row); err != nil {
			return err
		}
	}
	return nil
}
