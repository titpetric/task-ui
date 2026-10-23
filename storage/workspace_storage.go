package storage

import (
	"context"
	"errors"

	"github.com/titpetric/task-ui/model"
)

// WorkspaceStorage keeps the workspaces: checkouts under the root whose
// folder is the namespace of the projects cloned inside them.
type WorkspaceStorage struct {
	db DB
}

// NewWorkspaceStorage returns the storage over a pool.
func NewWorkspaceStorage(db DB) *WorkspaceStorage {
	return &WorkspaceStorage{db: db}
}

// Create inserts a workspace. The folder is its identity; any number
// of workspaces may share a repository. One without a sort goes after
// the others.
func (s *WorkspaceStorage) Create(ctx context.Context, workspace *model.Workspace) error {
	if workspace.Folder == "" {
		return errors.New("a workspace needs a folder")
	}
	if workspace.ID == "" {
		workspace.ID = model.NewID()
	}
	if workspace.Sort == 0 {
		next, err := client(s.db).Get[int64](ctx, "SELECT COALESCE(MAX(sort) + 1, 0) FROM workspace")
		if err != nil {
			return err
		}
		workspace.Sort = *next
	}
	workspace.CreatedAt = stamp()
	workspace.UpdatedAt = workspace.CreatedAt
	return client(s.db).Insert(ctx, model.WorkspaceTable, workspace)
}

// Get returns a workspace by ID.
func (s *WorkspaceStorage) Get(ctx context.Context, id string) (*model.Workspace, error) {
	workspace, err := client(s.db).Get[model.Workspace](ctx, "SELECT * FROM workspace WHERE id = ?", id)
	return workspace, notFound(err)
}

// GetByFolder returns the workspace of a folder.
func (s *WorkspaceStorage) GetByFolder(ctx context.Context, folder string) (*model.Workspace, error) {
	workspace, err := client(s.db).Get[model.Workspace](ctx, "SELECT * FROM workspace WHERE folder = ?", folder)
	return workspace, notFound(err)
}

// EnsureDefault adds the default workspace, whose directory is the
// root, when the row is missing. A default workspace already there keeps
// its row unchanged.
func (s *WorkspaceStorage) EnsureDefault(ctx context.Context) error {
	if _, err := s.GetByFolder(ctx, model.Default); !errors.Is(err, ErrNotFound) {
		return err
	}
	return s.Create(ctx, &model.Workspace{Folder: model.Default, Title: "Default"})
}

// List returns every workspace: the default first, then by sort, then by folder.
func (s *WorkspaceStorage) List(ctx context.Context) ([]model.Workspace, error) {
	return client(s.db).Select[model.Workspace](ctx, "SELECT * FROM workspace ORDER BY CASE WHEN folder = ? THEN 0 ELSE 1 END, sort, folder", model.Default)
}

// Folders returns the folder of every named workspace, in List's order,
// without the default workspace: its projects are the root's.
func (s *WorkspaceStorage) Folders(ctx context.Context) ([]string, error) {
	workspaces, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	folders := make([]string, 0, len(workspaces))
	for _, workspace := range workspaces {
		if workspace.Folder != model.Default {
			folders = append(folders, workspace.Folder)
		}
	}
	return folders, nil
}

// Update writes a workspace back, by ID.
func (s *WorkspaceStorage) Update(ctx context.Context, workspace *model.Workspace) error {
	workspace.UpdatedAt = stamp()
	return client(s.db).Update(ctx, model.WorkspaceTable, workspace, "id")
}

// Delete removes a workspace by ID, leaving its project rows to the caller.
func (s *WorkspaceStorage) Delete(ctx context.Context, id string) error {
	return client(s.db).Exec(ctx, "DELETE FROM workspace WHERE id = ?", id)
}
