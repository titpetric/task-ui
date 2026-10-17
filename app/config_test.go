package app_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestNewConfig(t *testing.T) {
	root := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(root, "task-ui.yml"), []byte("users: []\n"), 0o644))
	t.Setenv("TASK_UI_ROOT", root)
	t.Setenv("TASK_UI_LISTEN", "127.0.0.1:1234")
	t.Setenv("TASK_UI_DATA", filepath.Join(root, "data"))
	t.Setenv("TASK_UI_SESSION_TTL", "1h")
	t.Setenv("TASK_UI_KEEP", "2m")
	t.Setenv("TASK_UI_DISCOVER", "false")
	t.Setenv("TASK_UI_AUTOLOGIN", "admin")

	cfg, err := app.NewConfig(nil)
	assert.NoError(t, err)
	assert.Equal(t, root, cfg.Root)
	assert.Equal(t, "127.0.0.1:1234", cfg.Listen)
	assert.Equal(t, filepath.Join(root, "task-ui.yml"), cfg.UsersFile)
	assert.Equal(t, filepath.Join(root, "data", "task-ui.db"), cfg.DB)
	assert.Equal(t, filepath.Join(root, "data", "recordings"), cfg.RecordingsDir())
	assert.Equal(t, filepath.Join(root, "data", "artifacts"), cfg.ArtifactsDir())
	assert.Equal(t, time.Hour, cfg.SessionTTL)
	assert.Equal(t, 2*time.Minute, cfg.Keep)
	assert.Equal(t, 10*time.Second, cfg.Grace)
	assert.False(t, cfg.Discover)
	assert.Equal(t, "admin", cfg.AutoLogin)

	// Flags win over the environment.
	flagged, err := app.NewConfig([]string{"-l", ":9000", "--root", root, "--config", "/x/users.yml", "--data", "/tmp/d", "-v"})
	assert.NoError(t, err)
	assert.Equal(t, ":9000", flagged.Listen)
	assert.Equal(t, "/x/users.yml", flagged.UsersFile)
	assert.Equal(t, "/tmp/d/task-ui.db", flagged.DB)
	assert.True(t, flagged.PrintBuildInfo)
}

func TestNewConfig_Errors(t *testing.T) {
	t.Setenv("TASK_UI_ROOT", filepath.Join(t.TempDir(), "missing"))
	_, err := app.NewConfig(nil)
	assert.ErrorContains(t, err, "not a directory")

	t.Setenv("TASK_UI_ROOT", t.TempDir())
	t.Setenv("TASK_UI_SESSION_TTL", "soon")
	_, err = app.NewConfig(nil)
	assert.ErrorContains(t, err, "TASK_UI_SESSION_TTL")

	t.Setenv("TASK_UI_SESSION_TTL", "")
	_, err = app.NewConfig([]string{"--nope"})
	assert.Error(t, err)
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	cfg := &app.Config{Root: t.TempDir(), DataDir: "d", SessionTTL: time.Hour}
	assert.NoError(t, cfg.Validate())
	assert.Equal(t, filepath.Join("d", "task-ui.db"), cfg.DB)
	assert.Equal(t, "", cfg.UsersFile)

	assert.ErrorContains(t, (&app.Config{}).Validate(), "root directory")
	assert.ErrorContains(t, (&app.Config{Root: t.TempDir()}).Validate(), "positive")
}
