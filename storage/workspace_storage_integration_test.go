//go:build integration

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

	workspace := &model.Workspace{Folder: "malta"}
	assert.NoError(t, workspaces.Create(ctx, workspace))
	assert.True(t, model.ValidID(workspace.ID))
	assert.NotNil(t, workspace.CreatedAt)

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

	workspace.Title, workspace.Description, workspace.SidebarLimit = "Malta", "the island", 3
	assert.NoError(t, workspaces.Update(ctx, workspace))
	got, err := workspaces.Get(ctx, workspace.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Malta", got.Title)
	assert.Equal(t, "the island", got.Description)
	assert.Equal(t, int64(3), got.SidebarLimit, "the sidebar limit round trips")
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

// EnsureDefault makes the default workspace on an empty table alone: an
// instance with a named workspace and no default one, which deleted it,
// stays without.
func TestWorkspaceStorage_EnsureDefault(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	named := storage.NewWorkspaceStorage(newDB(t))
	assert.NoError(t, named.Create(ctx, &model.Workspace{Folder: "malta"}))
	assert.NoError(t, named.EnsureDefault(ctx))
	_, err := named.GetByFolder(ctx, model.Default)
	assert.ErrorIs(t, err, storage.ErrNotFound, "a named workspace alone brings no default one")

	workspaces := storage.NewWorkspaceStorage(newDB(t))
	assert.NoError(t, workspaces.EnsureDefault(ctx))
	def, err := workspaces.GetByFolder(ctx, model.Default)
	assert.NoError(t, err)
	assert.Equal(t, "Default", def.Title)
	assert.NoError(t, workspaces.Create(ctx, &model.Workspace{Folder: "malta"}))

	// Again, it keeps the row and its title.
	def.Title = "The root"
	assert.NoError(t, workspaces.Update(ctx, def))
	assert.NoError(t, workspaces.EnsureDefault(ctx))
	list, err := workspaces.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	// The default comes first whatever its sort, and Folders leaves it out.
	assert.Equal(t, model.Default, list[0].Folder)
	assert.Equal(t, def.ID, list[0].ID)
	assert.Equal(t, "The root", list[0].Title)
	folders, err := workspaces.Folders(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"malta"}, folders)
}

// A workspace's home repository, its URL and branch, is a column of the
// row: Create and Update keep it, and an empty URL is an empty home.
func TestWorkspaceStorage_Home(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	workspaces := storage.NewWorkspaceStorage(db)
	workspace := &model.Workspace{Folder: "malta", HomeURL: "git@example.com:org/workspace.git", HomeBranch: "main"}
	assert.NoError(t, workspaces.Create(ctx, workspace))

	got, err := workspaces.GetByFolder(ctx, "malta")
	assert.NoError(t, err)
	assert.Equal(t, "git@example.com:org/workspace.git", got.HomeURL)
	assert.Equal(t, "main", got.HomeBranch)
	assert.Equal(t, "", got.HomeProjectID, "the old column stays empty")

	got.HomeURL, got.HomeBranch = "", ""
	assert.NoError(t, workspaces.Update(ctx, got))
	got, err = workspaces.Get(ctx, workspace.ID)
	assert.NoError(t, err)
	assert.Equal(t, "", got.HomeURL, "Remove leaves an empty home")
	assert.Equal(t, "", got.HomeBranch)
}
