//go:build integration

package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestCatalogue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	makeFolder(t, root, "site")
	makeFolder(t, root, "infra")

	projects := storage.NewProjectStorage(db)
	var c folder.Catalogue = app.NewCatalogue(root, true, projects)
	list := c.List()
	assert.Len(t, list, 2)
	assert.Equal(t, "infra", list[0].ID)
	assert.Nil(t, c.Find(model.Internal), "no internal folder")
	assert.Equal(t, "site", c.Find("site").ID)
	assert.Nil(t, c.Find("nowhere"))
	// _ is the root, for its built-in jobs, and lists as no folder;
	// with no default workspace row, the root itself.
	rootFolder := c.Find(model.Default)
	assert.Equal(t, root, rootFolder.Dir)
	assert.NotNil(t, rootFolder.Job("@host:dev1"))

	// A project row changes what the folder says, on the next call.
	assert.NoError(t, projects.Create(ctx, &model.Project{Folder: "infra", Title: "The servers", Allow: model.EncodeList([]string{"ops"})}))
	infra := c.Find("infra")
	assert.Equal(t, "The servers", infra.Title)
	assert.Equal(t, []string{"ops"}, infra.Allow)

	// Without discovery the catalogue lists only configured folders.
	narrow := app.NewCatalogue(root, false, projects)
	assert.Len(t, narrow.List(), 1)
	assert.Equal(t, "infra", narrow.List()[0].ID)

	bare := app.NewCatalogue(root, true, nil)
	assert.Len(t, bare.List(), 2)
}

// A shallow project row cuts the catalogue at the project's top-level
// folders, on the next call: a deeper path resolves to the folder above
// it, whose page browses the directory.
func TestCatalogue_Shallow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	makeFolder(t, root, "mono")
	makeFolder(t, root, "mono/blog")
	makeFolder(t, root, "mono/blog/admin")

	projects := storage.NewProjectStorage(db)
	c := app.NewCatalogue(root, true, projects)
	assert.NotNil(t, c.Find("mono/blog/admin"))

	assert.NoError(t, projects.Create(ctx, &model.Project{Folder: "mono", Shallow: true}))
	assert.NotNil(t, c.Find("mono/blog"))
	assert.Nil(t, c.Find("mono/blog/admin"))
	assert.NotNil(t, c.Steps().Find("mono/blog/admin"), "hidden, not gone")
	f, rest := folder.Resolve(c, "mono/blog/admin")
	assert.Equal(t, "mono/blog", f.ID)
	assert.Equal(t, "admin", rest)
}

// No folder is internal: a directory or a project row of the name
// lists nothing, as the folder of older runs.
func TestCatalogue_internal(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	makeFolder(t, root, "internal")
	makeFolder(t, root, "internal/tools")
	projects := storage.NewProjectStorage(db)
	assert.NoError(t, projects.Create(ctx, &model.Project{Folder: "internal/tools", Title: "Tools"}))

	c := app.NewCatalogue(root, true, projects)
	assert.Nil(t, c.Find(model.Internal))
	assert.Nil(t, c.Find("internal/tools"))
	assert.Len(t, c.List(), 0)
}

// makeFolder puts a Makefile with one target in a folder of the root's src.
func makeFolder(t *testing.T, root, id string) {
	t.Helper()
	dir := filepath.Join(root, layout.Src, id)
	assert.NoError(t, os.MkdirAll(dir, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "Makefile"), []byte("build:\n\techo built\n"), 0o644))
}

// A directory under a workspace has the workspace's home as the HOME of
// its runs, whether the home holds a checkout or not; one outside the
// root has none.
func TestCatalogue_Home(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	projects, workspaces := storage.NewProjectStorage(db), storage.NewWorkspaceStorage(db)
	assert.NoError(t, workspaces.Create(ctx, &model.Workspace{Folder: "malta", HomeURL: "git@example.com:org/workspace.git"}))
	assert.NoError(t, workspaces.Create(ctx, &model.Workspace{Folder: "gozo"}))
	assert.NoError(t, workspaces.Create(ctx, &model.Workspace{Folder: model.Default, Title: "Default"}))

	c := app.NewCatalogue(root, true, projects)
	c.Workspaces = workspaces
	want := filepath.Join(root, "malta", "home")
	assert.Equal(t, want, c.Home(filepath.Join(root, "malta", "src", "dockerfiles")))
	assert.Equal(t, want, c.Home(filepath.Join(root, "malta", "home", "scripts")))
	assert.Equal(t, filepath.Join(root, "gozo", "home"), c.Home(filepath.Join(root, "gozo", "src", "api")), "an empty home is a home")
	assert.Equal(t, filepath.Join(root, "home"), c.Home(filepath.Join(root, "src", "deploy")), "the default workspace's home is at the root")
	assert.Equal(t, "", c.Home(t.TempDir()))
}
