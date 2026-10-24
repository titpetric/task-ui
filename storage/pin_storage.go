package storage

import (
	"context"
	"errors"

	"github.com/titpetric/task-ui/model"
)

// PinStorage keeps what each user pins to the sidebar: folders and jobs,
// one row per user and target, the user's own.
type PinStorage struct {
	db DB
}

// NewPinStorage returns the storage over a pool.
func NewPinStorage(db DB) *PinStorage {
	return &PinStorage{db: db}
}

// List returns a user's pins, oldest first.
func (s *PinStorage) List(ctx context.Context, user string) ([]model.Pin, error) {
	pins, err := client(s.db).Select[model.Pin](ctx, "SELECT * FROM pin WHERE user_name = ? ORDER BY rowid", user)
	if pins == nil {
		pins = []model.Pin{}
	}
	return pins, err
}

// Add stores a pin, or fills it from the user's pin of the same target.
func (s *PinStorage) Add(ctx context.Context, pin *model.Pin) error {
	if pin.UserName == "" || pin.Folder == "" {
		return errors.New("a pin needs a user and a folder")
	}
	if pin.Args == "" {
		pin.Args = "[]"
	}
	had, err := client(s.db).Get[model.Pin](ctx, "SELECT * FROM pin WHERE user_name = ? AND folder = ? AND job = ? AND dir = ? AND args = ?", pin.UserName, pin.Folder, pin.Job, pin.Dir, pin.Args)
	if err = notFound(err); err == nil {
		*pin = *had
		return nil
	}
	if !errors.Is(err, ErrNotFound) {
		return err
	}
	pin.ID, pin.CreatedAt = model.NewID(), stamp()
	return client(s.db).Insert(ctx, model.PinTable, pin)
}

// Delete removes a user's pin by ID, ErrNotFound for none of the user's.
func (s *PinStorage) Delete(ctx context.Context, user, id string) error {
	if _, err := client(s.db).Get[model.Pin](ctx, "SELECT * FROM pin WHERE id = ? AND user_name = ?", id, user); err != nil {
		return notFound(err)
	}
	return client(s.db).Exec(ctx, "DELETE FROM pin WHERE id = ?", id)
}
