package app

import (
	"context"
	"log/slog"

	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/retention"
	"github.com/titpetric/task-ui/scheduler"
	"github.com/titpetric/task-ui/storage"
)

// NewCleaner returns the cleaner of the data the configuration names,
// over an open database: what task-ui cleanup runs, and the Data
// retention page measures.
func NewCleaner(db storage.DB, cfg *Config) *retention.Cleaner {
	return &retention.Cleaner{
		DB: db, DBPath: cfg.DB,
		Values:        storage.NewSettingStorage(db),
		Runs:          storage.NewRunStorage(db),
		Artifacts:     storage.NewArtifactStorage(db),
		Store:         &artifactstore.Store{Dir: cfg.ArtifactTree(), Keep: cfg.ArtifactHistory},
		SecretReads:   storage.NewSecretReadStorage(db),
		KeyReads:      storage.NewKeyReadStorage(db),
		RecordingsDir: cfg.RecordingsDir(),
		Now:           cfg.Clock(),
	}
}

// seededCleanup is the setting that says the @cleanup schedule went in.
const seededCleanup = "cleanup.seeded"

// seedCleanup adds the daily schedule of the root's @cleanup once, run
// as the auto login user when an admin, else the first admin by name;
// until there is one it waits for the next start. An admin edits or deletes it on the
// Schedules page, and it stays as they leave it.
func (a *App) seedCleanup(ctx context.Context) error {
	settings := storage.NewSettingStorage(a.DB)
	values, err := settings.Values(ctx)
	if err != nil || values[seededCleanup] != "" {
		return err
	}
	var admin *model.User
	if user, err := a.Users.GetByName(ctx, a.Config.AutoLogin); err == nil && user.CanAdmin() {
		admin = user
	}
	if users, err := a.Users.List(ctx); err == nil && admin == nil {
		for i := range users {
			if users[i].CanAdmin() {
				admin = &users[i]
				break
			}
		}
	}
	if admin == nil {
		return nil
	}
	next, err := scheduler.NextRun("@daily", a.Config.Clock()())
	if err != nil {
		return err
	}
	schedule := &model.Schedule{Folder: model.Default, Job: folder.CleanupID, UserID: admin.ID, UserName: admin.Name, Expression: "@daily", Enabled: 1, NextRunAt: &next}
	if err := storage.NewScheduleStorage(a.DB).Create(ctx, schedule); err != nil {
		return err
	}
	slog.Info("scheduled the data retention cleanup", "job", model.Default+"/"+folder.CleanupID, "expression", "@daily", "user", admin.Name)
	return settings.Set(ctx, map[string]string{seededCleanup: schedule.ID})
}
