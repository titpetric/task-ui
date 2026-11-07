package api

import (
	"context"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// RunStore keeps the runs the handlers read: one by ID, a listing, the
// newest run a filter matches and the ones going; storage.RunStorage, or
// a fake in a test.
type RunStore interface {
	Get(ctx context.Context, id string) (*model.Run, error)
	List(ctx context.Context, filter storage.RunFilter) ([]model.Run, error)
	Last(ctx context.Context, filter storage.RunFilter) (*model.Run, error)
	Active(ctx context.Context) ([]model.Run, error)
}
