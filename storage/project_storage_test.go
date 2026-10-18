package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestProjectStorage_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	projects := storage.NewProjectStorage(newDB(t))

	project := &model.Project{Folder: "infra", Title: "Infrastructure"}
	assert.NoError(t, projects.Create(ctx, project))
	assert.True(t, model.ValidID(project.ID))
	assert.Equal(t, "[]", project.Allow)
	assert.Equal(t, "{}", project.Commands)

	assert.Error(t, projects.Create(ctx, &model.Project{}), "no folder")
	assert.Error(t, projects.Create(ctx, &model.Project{Folder: "infra"}), "duplicate folder")
}

func TestProjectStorage_Get(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	projects := storage.NewProjectStorage(newDB(t))
	project := &model.Project{Folder: "infra", Allow: model.EncodeList([]string{"ops"}), JobAllow: model.EncodeMap(map[string][]string{"deploy": {"admin"}})}
	assert.NoError(t, projects.Create(ctx, project))

	got, err := projects.Get(ctx, project.ID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"ops"}, got.AllowList())
	assert.Equal(t, []string{"admin"}, got.JobAllowMap()["deploy"])

	byFolder, err := projects.GetByFolder(ctx, "infra")
	assert.NoError(t, err)
	assert.Equal(t, project.ID, byFolder.ID)

	_, err = projects.Get(ctx, "missing")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = projects.GetByFolder(ctx, "missing")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestProjectStorage_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	projects := storage.NewProjectStorage(newDB(t))
	for _, folder := range []string{"web", "infra"} {
		assert.NoError(t, projects.Create(ctx, &model.Project{Folder: folder}))
	}
	list, err := projects.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "infra", list[0].Folder)
}

func TestProjectStorage_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	projects := storage.NewProjectStorage(newDB(t))
	project := &model.Project{Folder: "infra"}
	assert.NoError(t, projects.Create(ctx, project))

	project.Title = "Infra"
	project.Commands = model.EncodeStrings(map[string]string{"make": "gmake"})
	assert.NoError(t, projects.Update(ctx, project))
	got, err := projects.Get(ctx, project.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Infra", got.Title)
	assert.Equal(t, "gmake", got.CommandMap()["make"])
}

func TestProjectStorage_Upsert(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	projects := storage.NewProjectStorage(newDB(t))

	first := &model.Project{Folder: "infra", Title: "One"}
	assert.NoError(t, projects.Upsert(ctx, first))
	second := &model.Project{Folder: "infra", Title: "Two"}
	assert.NoError(t, projects.Upsert(ctx, second))
	assert.Equal(t, first.ID, second.ID, "the row keeps its id")

	list, err := projects.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Two", list[0].Title)
}

func TestProjectStorage_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	projects := storage.NewProjectStorage(newDB(t))
	project := &model.Project{Folder: "infra"}
	assert.NoError(t, projects.Create(ctx, project))
	assert.NoError(t, projects.Delete(ctx, project.ID))
	_, err := projects.Get(ctx, project.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
}
