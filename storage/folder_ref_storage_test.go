package storage_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

func TestFolderRefStorage_Under(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	refs := storage.NewFolderRefStorage(db)

	got, err := refs.Under(ctx, "web")
	assert.NoError(t, err)
	assert.Len(t, got, 0)

	schedules := storage.NewScheduleStorage(db)
	assert.NoError(t, schedules.Create(ctx, &model.Schedule{Folder: "web", Job: "build", Expression: "@hourly"}))
	assert.NoError(t, schedules.Create(ctx, &model.Schedule{Folder: "website", Job: "build", Expression: "@hourly"}))
	assert.NoError(t, schedules.Create(ctx, &model.Schedule{Folder: "w_b", Job: "build", Expression: "@hourly"}))
	secrets := storage.NewSecretStorage(db)
	assert.NoError(t, secrets.Create(ctx, &model.Secret{Scope: "web/docs", Name: "TOKEN", Content: "x"}))
	assert.NoError(t, secrets.Create(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "SHARED", Content: "x"}))
	files := storage.NewEnvFileStorage(db)
	assert.NoError(t, files.Create(ctx, &model.EnvFile{Folder: "web", Path: ".env"}))
	workflows := storage.NewWorkflowStorage(db)
	assert.NoError(t, workflows.Create(ctx, &model.Workflow{Name: "release"}, []model.WorkflowStep{{Folder: "web", Job: "deploy"}}))

	got, err = refs.Under(ctx, "web")
	assert.NoError(t, err)
	assert.Len(t, got, 4)
	assert.Equal(t, "env file web · .env", got[0].String())
	assert.Equal(t, "schedule web · build", got[1].String())
	assert.Equal(t, "secret web/docs · TOKEN", got[2].String())
	assert.Equal(t, "workflow step web · release: deploy", got[3].String())

	got, err = refs.Under(ctx, "w_b")
	assert.NoError(t, err)
	assert.Len(t, got, 1, "an _ in the ID matches itself alone")
}
