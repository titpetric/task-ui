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
// of workspaces may share a repository.
func (s *WorkspaceStorage) Create(ctx context.Context, workspace *model.Workspace) error {
	if workspace.Folder == "" {
		return errors.New("a workspace needs a folder")
	}
	if workspace.ID == "" {
		workspace.ID = model.NewID()
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

// List returns every workspace, by folder.
func (s *WorkspaceStorage) List(ctx context.Context) ([]model.Workspace, error) {
	return client(s.db).Select[model.Workspace](ctx, "SELECT * FROM workspace ORDER BY folder")
}

// Folders returns the folder of every workspace, sorted.
func (s *WorkspaceStorage) Folders(ctx context.Context) ([]string, error) {
	workspaces, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	folders := make([]string, 0, len(workspaces))
	for _, workspace := range workspaces {
		folders = append(folders, workspace.Folder)
	}
	return folders, nil
}

// Update writes a workspace back, by ID.
func (s *WorkspaceStorage) Update(ctx context.Context, workspace *model.Workspace) error {
	workspace.UpdatedAt = stamp()
	return client(s.db).Update(ctx, model.WorkspaceTable, workspace, "id")
}

// Delete removes a workspace by ID. The project rows inside it are the
// caller's to remove.
func (s *WorkspaceStorage) Delete(ctx context.Context, id string) error {
	return client(s.db).Exec(ctx, "DELETE FROM workspace WHERE id = ?", id)
}
