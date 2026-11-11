package envfile

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// Rows keeps the content users fill for env files, one row per folder
// and path.
type Rows interface {
	Get(ctx context.Context, ref model.EnvFile) (*model.EnvFile, error)
	List(ctx context.Context, folder string) ([]model.EnvFile, error)
	All(ctx context.Context) ([]model.EnvFile, error)
	Set(ctx context.Context, file *model.EnvFile) error
	Update(ctx context.Context, file *model.EnvFile) error
	Delete(ctx context.Context, id string) error
}
