package forgejo

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// Workspaces are the workspace rows sync reads and writes, as
// storage.WorkspaceStorage keeps them.
type Workspaces interface {
	GetByFolder(ctx context.Context, folder string) (*model.Workspace, error)
	Create(ctx context.Context, workspace *model.Workspace) error
	Update(ctx context.Context, workspace *model.Workspace) error
}
