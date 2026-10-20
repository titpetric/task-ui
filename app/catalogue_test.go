package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite" // the driver storage opens

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// makeFolder puts a Makefile with one target under root.
func makeFolder(t *testing.T, root, id string) {
	t.Helper()
	assert.NoError(t, os.MkdirAll(filepath.Join(root, id), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(root, id, "Makefile"), []byte("build:\n\techo built\n"), 0o644))
}

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
	assert.Len(t, list, 3)
	assert.Equal(t, model.Internal, list[0].ID, "the internal folder comes first")
	assert.Equal(t, "infra", list[1].ID)
	assert.Equal(t, "site", c.Find("site").ID)
	assert.Nil(t, c.Find("nowhere"))

	// A project row changes what the folder says, on the next call.
	assert.NoError(t, projects.Create(ctx, &model.Project{Folder: "infra", Title: "The servers", Allow: model.EncodeList([]string{"ops"})}))
	infra := c.Find("infra")
	assert.Equal(t, "The servers", infra.Title)
	assert.Equal(t, []string{"ops"}, infra.Allow)

	// Without discovery the catalogue lists only configured folders.
	narrow := app.NewCatalogue(root, false, projects)
	assert.Len(t, narrow.List(), 2)
	assert.Equal(t, "infra", narrow.List()[1].ID)

	bare := app.NewCatalogue(root, true, nil)
	assert.Len(t, bare.List(), 3)
}

// The internal folder holds the host commands, with the agent's user on
// the notes job; no directory or project row takes its place.
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

	c := &app.Catalogue{Root: root, Discover: true, Projects: projects, Agent: "claude"}
	internal := c.Find(model.Internal)
	assert.NotNil(t, internal)
	assert.Equal(t, root, internal.Dir)
	assert.Len(t, internal.Jobs, 3)
	assert.Equal(t, []string{"claude"}, internal.Job("claude-note").Allow)
	assert.Equal(t, []string{"admin"}, internal.Job("claude-prompt").Allow)
	assert.True(t, internal.Job("claude").Interactive)
	assert.Nil(t, c.Find("internal/tools"))

	list := c.List()
	assert.Len(t, list, 1)
	assert.Equal(t, model.Internal, list[0].ID)
}
