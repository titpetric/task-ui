package storage

import (
	"context"
	"errors"

	"github.com/titpetric/task-ui/app/model"
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

// movedPins matches the pins of a folder and of the folders below that
// folder. ?4 || substr(folder, ?5) is where each pin moves to: the new
// folder with what followed the old one. It is model.Moved as a column
// expression, and the arguments are movedArgs.
const movedPins = "folder = ?1 OR substr(folder, 1, ?2) = ?3"

// movedArgs are the arguments of movedPins and of the expression that
// rewrites a folder: the old folder, the length of its prefix and the
// prefix, then the new folder and where the suffix of the old one starts.
func movedArgs(from, to string) []any {
	prefix := from + "/"
	return []any{from, len(prefix), prefix, to, len(from) + 1}
}

// MoveFolder renames the pins of folder from, and of the folders below
// from, to the ID to, for every user, and returns how many moved. A pin
// whose user holds the target already goes, the target's pin stays.
func (s *PinStorage) MoveFolder(ctx context.Context, from, to string) (int, error) {
	db := client(s.db)
	if err := db.Exec(ctx, `DELETE FROM pin WHERE (`+movedPins+`)
		AND EXISTS (SELECT 1 FROM pin AS held WHERE held.id <> pin.id
			AND held.user_name = pin.user_name AND held.folder = ?4 || substr(pin.folder, ?5)
			AND held.job = pin.job AND held.dir = pin.dir AND held.args = pin.args)`,
		movedArgs(from, to)...); err != nil {
		return 0, err
	}
	if err := db.Exec(ctx, "UPDATE pin SET folder = ?4 || substr(folder, ?5) WHERE "+movedPins, movedArgs(from, to)...); err != nil {
		return 0, err
	}
	return int(db.RowsAffected()), nil
}
