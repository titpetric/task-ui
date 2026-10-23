//go:build integration

package app_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// The first start schedules @cleanup daily as the first admin, once: a
// second start over the same database adds none, nor after an admin
// deleted it.
func TestNew_SeedsCleanup(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	makeFolder(t, root, "site")
	data := filepath.Join(t.TempDir(), "data")
	cfg := &app.Config{
		Listen: "127.0.0.1:0", Root: root, Htpasswd: htpasswdFile(t, "admin", "guest"), Import: &app.RoleImport{Admins: []string{"nobody", "admin"}},
		DataDir: data, SSHDir: t.TempDir(), SessionTTL: time.Hour, Keep: time.Minute, Grace: time.Second,
	}
	assert.NoError(t, cfg.Validate())
	ctx := context.Background()
	start := func() []model.Schedule {
		a, err := app.New(ctx, cfg, "test")
		assert.NoError(t, err)
		defer a.Close()
		schedules, err := storage.NewScheduleStorage(a.DB).List(ctx)
		assert.NoError(t, err)
		return schedules
	}
	schedules := start()
	assert.Len(t, schedules, 1)
	assert.Equal(t, model.Default, schedules[0].Folder)
	assert.Equal(t, folder.CleanupID, schedules[0].Job)
	assert.Equal(t, "@daily", schedules[0].Expression)
	assert.Equal(t, "admin", schedules[0].UserName)
	assert.Equal(t, 0, schedules[0].NextRunAt.Local().Hour(), "next at midnight")

	assert.Len(t, start(), 1)

	db, err := storage.OpenAndMigrate(ctx, cfg.DB)
	assert.NoError(t, err)
	assert.NoError(t, storage.NewScheduleStorage(db).Delete(ctx, schedules[0].ID))
	db.Close()
	assert.Len(t, start(), 0)
}

// The cleaner works on the data the configuration names.
func TestNewCleaner(t *testing.T) {
	t.Parallel()
	cfg := &app.Config{DataDir: t.TempDir(), DB: ":memory:", ArtifactHistory: 7}
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	defer db.Close()
	c := app.NewCleaner(db, cfg)
	assert.Equal(t, cfg.RecordingsDir(), c.RecordingsDir)
	assert.Equal(t, cfg.ArtifactTree(), c.Store.Dir)
	assert.Equal(t, 7, c.Store.Keep)
}
