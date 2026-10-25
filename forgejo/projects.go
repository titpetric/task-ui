package forgejo

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// Projects are the project rows sync reads and writes, as
// storage.ProjectStorage keeps them.
type Projects interface {
	GetByFolder(ctx context.Context, folder string) (*model.Project, error)
	Upsert(ctx context.Context, project *model.Project) error
	List(ctx context.Context) ([]model.Project, error)
}
