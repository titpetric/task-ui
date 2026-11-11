package api

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// PinStore keeps the pins of each user, what /pins lists, adds and
// removes, and what a project that moves takes along: storage.PinStorage,
// or a fake in a test.
type PinStore interface {
	List(ctx context.Context, user string) ([]model.Pin, error)
	Add(ctx context.Context, pin *model.Pin) error
	Delete(ctx context.Context, user, id string) error
	// MoveFolder renames the pins of a folder, and of the folders inside
	// it, to its new ID, for every user, and counts them.
	MoveFolder(ctx context.Context, from, to string) (int, error)
}
