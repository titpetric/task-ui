package app

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/retention"
	"github.com/titpetric/task-ui/scheduler"
	"github.com/titpetric/task-ui/storage"
)

// NewCleaner returns the cleaner over db; serverHome holds the instance's own tool history.
func NewCleaner(db storage.DB, cfg *Config, serverHome string) *retention.Cleaner {
	return &retention.Cleaner{
		DB: db, DBPath: cfg.DB,
		Values:        storage.NewSettingStorage(db),
		Runs:          storage.NewRunStorage(db),
		Summaries:     storage.NewRunSummaryStorage(db),
		Notes:         storage.NewNoteStorage(db),
		Flows:         storage.NewWorkflowRunStorage(db),
		Artifacts:     storage.NewArtifactStorage(db),
		Store:         &artifactstore.Store{Dir: cfg.ArtifactTree(), Keep: cfg.ArtifactHistory},
		SecretReads:   storage.NewSecretReadStorage(db),
		Tokens:        storage.NewUserAuthTokenStorage(db),
		Sessions:      storage.NewUserSessionStorage(db),
		RecordingsDir: cfg.RecordingsDir(),
		Tooling:       toolHistory(cfg, storage.NewWorkspaceStorage(db), serverHome),
		Atkins:        atkins(),
		Self:          os.Getenv("TASK_UI_RUN_ID"),
		Now:           cfg.Clock(),
	}
}

// toolHistory returns the atkins.db files the Tool history row measures:
// each workspace's home's and the instance's own under the server's HOME.
// Every measure reads the workspaces again.
func toolHistory(cfg *Config, workspaces *storage.WorkspaceStorage, serverHome string) func() []string {
	return func() []string {
		var files []string
		rows, _ := workspaces.List(context.Background())
		for _, row := range rows {
			files = append(files, filepath.Join(layout.HomeDir(cfg.Root, row.Folder), ".atkins", "atkins.db"))
		}
		if serverHome != "" {
			files = append(files, filepath.Join(serverHome, ".atkins", "atkins.db"))
		}
		return files
	}
}

// atkins returns a run of the atkins binary on PATH, which the Tool
// history row prunes through; nil, said once in the log, when there is none.
func atkins() func(ctx context.Context, args ...string) ([]byte, error) {
	path, err := exec.LookPath("atkins")
	if err != nil {
		slog.Warn("atkins not on PATH: the cleanup measures the tool history and prunes none of it")
		return nil
	}
	return func(ctx context.Context, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, path, args...).CombinedOutput()
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
