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
	assert.NoError(t, os.WriteFile(filepath.Join(root, ".htpasswd"), []byte(""), 0o600))
	t.Setenv("TASK_UI_ADMINS", "admin, root,")
	t.Setenv("TASK_UI_VIEWERS", "guest")
	t.Setenv("TASK_UI_ROOT", root)
	t.Setenv("TASK_UI_LISTEN", "127.0.0.1:1234")
	t.Setenv("TASK_UI_DATA", filepath.Join(root, "data"))
	t.Setenv("TASK_UI_SESSION_TTL", "1h")
	t.Setenv("TASK_UI_KEEP", "2m")
	t.Setenv("TASK_UI_DISCOVER", "false")
	t.Setenv("TASK_UI_AUTOLOGIN", "admin")
	t.Setenv("TASK_UI_NOW", "2026-10-16T22:00:00+02:00")

	cfg, err := app.NewConfig(nil)
	assert.NoError(t, err)
	assert.Equal(t, root, cfg.Root)
	assert.Equal(t, "127.0.0.1:1234", cfg.Listen)
	assert.Equal(t, filepath.Join(root, ".htpasswd"), cfg.Htpasswd)
	assert.Equal(t, []string{"admin", "root"}, cfg.Admins)
	assert.Equal(t, []string{"guest"}, cfg.Viewers)
	assert.Equal(t, filepath.Join(root, "data", "task-ui.db"), cfg.DB)
	assert.Equal(t, filepath.Join(root, "data", "recordings"), cfg.RecordingsDir())
	assert.Equal(t, filepath.Join(root, "data", "artifacts"), cfg.ArtifactsDir())
	assert.Equal(t, time.Hour, cfg.SessionTTL)
	assert.Equal(t, 2*time.Minute, cfg.Keep)
	assert.Equal(t, 10*time.Second, cfg.Grace)
	assert.False(t, cfg.Discover)
	assert.Equal(t, "admin", cfg.AutoLogin)
	assert.Equal(t, 22, cfg.Clock()().Hour())
	assert.Equal(t, time.Friday, cfg.Clock()().Weekday())

	// Flags win over the environment.
	flagged, err := app.NewConfig([]string{"-l", ":9000", "--root", root, "--htpasswd", "/x/.htpasswd", "--data", "/tmp/d", "-v"})
	assert.NoError(t, err)
	assert.Equal(t, ":9000", flagged.Listen)
	assert.Equal(t, "/x/.htpasswd", flagged.Htpasswd)
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

	t.Setenv("TASK_UI_NOW", "yesterday")
	_, err = app.NewConfig(nil)
	assert.ErrorContains(t, err, "TASK_UI_NOW")
	t.Setenv("TASK_UI_NOW", "")
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	cfg := &app.Config{Root: t.TempDir(), DataDir: "d", SessionTTL: time.Hour}
	assert.NoError(t, cfg.Validate())
	assert.Equal(t, filepath.Join("d", "task-ui.db"), cfg.DB)
	assert.Equal(t, "", cfg.Htpasswd)
	assert.Nil(t, cfg.Admins)
	assert.True(t, time.Since(cfg.Clock()()) < time.Minute)

	assert.ErrorContains(t, (&app.Config{}).Validate(), "root directory")
	assert.ErrorContains(t, (&app.Config{Root: t.TempDir()}).Validate(), "positive")
}
