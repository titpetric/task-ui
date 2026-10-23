package secret

import (
	"context"
	"fmt"
	"strings"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Workspaces finds the workspace a secret's scope belongs to.
// workspace:<folder> names its workspace; a folder's is its first
// segment when that is a named workspace, else the default one, _.
type Workspaces struct {
	Rows *storage.WorkspaceStorage
}

// Service reports whether a scope is a service's, such as
// integration:<kind>, encrypted with the instance key: a scope with a
// colon that is not a workspace's. No folder ID has a colon.
func Service(scope string) bool {
	return strings.Contains(scope, ":") && !strings.HasPrefix(scope, storage.ScopeWorkspace)
}

// Name returns the folder of the workspace a scope belongs to. The
// scope all, the shared secrets of an instance that predates
// workspaces, is the default workspace's.
func (w *Workspaces) Name(ctx context.Context, scope string) (string, error) {
	if name, ok := strings.CutPrefix(scope, storage.ScopeWorkspace); ok {
		return name, nil
	}
	if scope == storage.ScopeAll {
		return model.Default, nil
	}
	top, _, _ := strings.Cut(scope, "/")
	if top == model.Default || w == nil || w.Rows == nil {
		return model.Default, nil
	}
	if _, err := w.Rows.GetByFolder(ctx, top); err == nil {
		return top, nil
	}
	return model.Default, nil
}

// Of returns the workspace row a scope belongs to.
func (w *Workspaces) Of(ctx context.Context, scope string) (*model.Workspace, error) {
	name, err := w.Name(ctx, scope)
	if err != nil {
		return nil, err
	}
	row, err := w.Rows.GetByFolder(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("the workspace %s of %s: %w", name, scope, err)
	}
	return row, nil
}
