package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestWorkspaceStorage_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	workspaces := storage.NewWorkspaceStorage(newDB(t))

	workspace := &model.Workspace{Folder: "malta", SourceURL: "git@example.com:malta/workspace.git"}
	assert.NoError(t, workspaces.Create(ctx, workspace))
	assert.True(t, model.ValidID(workspace.ID))
	assert.NotNil(t, workspace.CreatedAt)

	same := &model.Workspace{Folder: "gozo", SourceURL: workspace.SourceURL}
	assert.NoError(t, workspaces.Create(ctx, same), "a repository may back more than one workspace")

	assert.Error(t, workspaces.Create(ctx, &model.Workspace{}), "no folder")
	assert.Error(t, workspaces.Create(ctx, &model.Workspace{Folder: "malta"}), "duplicate folder")
}

func TestWorkspaceStorage_Get(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	workspaces := storage.NewWorkspaceStorage(newDB(t))
	workspace := &model.Workspace{Folder: "malta", Title: "Malta"}
	assert.NoError(t, workspaces.Create(ctx, workspace))

	got, err := workspaces.Get(ctx, workspace.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Malta", got.Title)

	byFolder, err := workspaces.GetByFolder(ctx, "malta")
	assert.NoError(t, err)
	assert.Equal(t, workspace.ID, byFolder.ID)

	_, err = workspaces.Get(ctx, "missing")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = workspaces.GetByFolder(ctx, "missing")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestWorkspaceStorage_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	workspaces := storage.NewWorkspaceStorage(newDB(t))
	for _, folder := range []string{"malta", "gozo"} {
		assert.NoError(t, workspaces.Create(ctx, &model.Workspace{Folder: folder}))
	}
	// An added workspace goes last.
	list, err := workspaces.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "malta", list[0].Folder)
	assert.Equal(t, int64(0), list[0].Sort)
	assert.Equal(t, int64(1), list[1].Sort)

	folders, err := workspaces.Folders(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"malta", "gozo"}, folders)

	// The sort orders them, the folder breaks a tie.
	list[1].Sort = 0
	assert.NoError(t, workspaces.Update(ctx, &list[1]))
	folders, err = workspaces.Folders(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"gozo", "malta"}, folders)
	list[0].Sort = -1
	assert.NoError(t, workspaces.Update(ctx, &list[0]))
	folders, err = workspaces.Folders(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"malta", "gozo"}, folders)
}

func TestWorkspaceStorage_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	workspaces := storage.NewWorkspaceStorage(newDB(t))
	workspace := &model.Workspace{Folder: "malta"}
	assert.NoError(t, workspaces.Create(ctx, workspace))

	workspace.Title, workspace.Description = "Malta", "the island"
	assert.NoError(t, workspaces.Update(ctx, workspace))
	got, err := workspaces.Get(ctx, workspace.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Malta", got.Title)
	assert.Equal(t, "the island", got.Description)
}

func TestWorkspaceStorage_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	workspaces := storage.NewWorkspaceStorage(newDB(t))
	workspace := &model.Workspace{Folder: "malta"}
	assert.NoError(t, workspaces.Create(ctx, workspace))
	assert.NoError(t, workspaces.Delete(ctx, workspace.ID))
	_, err := workspaces.Get(ctx, workspace.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
}
