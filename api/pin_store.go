package api

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// PinStore keeps the pins of each user, what /pins lists, adds and
// removes: storage.PinStorage, or a fake in a test.
type PinStore interface {
	List(ctx context.Context, user string) ([]model.Pin, error)
	Add(ctx context.Context, pin *model.Pin) error
	Delete(ctx context.Context, user, id string) error
}
