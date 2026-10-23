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
	t.Setenv("TASK_UI_ADMINS", "admin, root,")
	t.Setenv("TASK_UI_VIEWERS", "guest")
	t.Setenv("TASK_UI_WORKSPACE", root)
	t.Setenv("TASK_UI_LISTEN", "127.0.0.1:1234")
	t.Setenv("TASK_UI_DATA", filepath.Join(root, "data"))
	t.Setenv("TASK_UI_SESSION_TTL", "1h")
	t.Setenv("TASK_UI_KEEP", "2m")
	t.Setenv("TASK_UI_ARTIFACT_HISTORY", "7")
	t.Setenv("TASK_UI_DISCOVER", "false")
	t.Setenv("TASK_UI_AUTOLOGIN", "admin")
	t.Setenv("TASK_UI_NOW", "2026-10-16T22:00:00+02:00")

	cfg, err := app.NewConfig(nil)
	assert.NoError(t, err)
	assert.Equal(t, root, cfg.Root)
	assert.Equal(t, "127.0.0.1:1234", cfg.Listen)
	assert.Equal(t, filepath.Join(root, "data", ".htpasswd"), cfg.Htpasswd, "in the data directory")
	assert.Equal(t, &app.RoleImport{Admins: []string{"admin", "root"}, Viewers: []string{"guest"}}, cfg.Import)
	assert.Equal(t, filepath.Join(root, "data", "task-ui.db"), cfg.DB)
	assert.Equal(t, filepath.Join(root, "data", "recordings"), cfg.RecordingsDir())
	assert.Equal(t, filepath.Join(root, "data", "artifacts"), cfg.ArtifactsDir())
	assert.Equal(t, filepath.Join(root, "data", "artifacts", "tree"), cfg.ArtifactTree())
	assert.Equal(t, 7, cfg.ArtifactHistory)
	assert.Equal(t, time.Hour, cfg.SessionTTL)
	assert.Equal(t, 2*time.Minute, cfg.Keep)
	assert.Equal(t, 10*time.Second, cfg.Grace)
	assert.False(t, cfg.Discover)
	assert.Equal(t, "admin", cfg.AutoLogin)
	assert.Equal(t, 22, cfg.Clock()().Hour())
	assert.Equal(t, time.Friday, cfg.Clock()().Weekday())

	// Flags win over the environment.
	flagged, err := app.NewConfig([]string{"-l", ":9000", "--workspace", root, "--data", "/tmp/d", "-v"})
	assert.NoError(t, err)
	assert.Equal(t, ":9000", flagged.Listen)
	assert.Equal(t, "/tmp/d/.htpasswd", flagged.Htpasswd)
	assert.Equal(t, "/tmp/d/task-ui.db", flagged.DB)
	assert.True(t, flagged.PrintBuildInfo)
}

// The role lists of an older configuration are an import, nil when
// none of them is in the environment.
func TestNewConfig_Import(t *testing.T) {
	t.Setenv("TASK_UI_WORKSPACE", t.TempDir())
	t.Setenv("TASK_UI_ADMINS", "")
	t.Setenv("TASK_UI_VIEWERS", "")
	t.Setenv("TASK_UI_AGENT", "")
	cfg, err := app.NewConfig(nil)
	assert.NoError(t, err)
	assert.Nil(t, cfg.Import)

	t.Setenv("TASK_UI_AGENT", "claude")
	cfg, err = app.NewConfig(nil)
	assert.NoError(t, err)
	assert.Equal(t, &app.RoleImport{Agent: "claude"}, cfg.Import)
}

// The base link drops a trailing slash; unset, the pages derive it per
// request.
func TestNewConfig_URL(t *testing.T) {
	t.Setenv("TASK_UI_WORKSPACE", t.TempDir())
	t.Setenv("TASK_UI_URL", "")
	cfg, err := app.NewConfig(nil)
	assert.NoError(t, err)
	assert.Equal(t, "", cfg.URL)

	t.Setenv("TASK_UI_URL", "https://task.example.com/")
	cfg, err = app.NewConfig(nil)
	assert.NoError(t, err)
	assert.Equal(t, "https://task.example.com", cfg.URL)
}

func TestNewConfig_Errors(t *testing.T) {
	t.Setenv("TASK_UI_WORKSPACE", filepath.Join(t.TempDir(), "missing"))
	_, err := app.NewConfig(nil)
	assert.ErrorContains(t, err, "not a directory")

	t.Setenv("TASK_UI_WORKSPACE", t.TempDir())
	t.Setenv("TASK_UI_SESSION_TTL", "soon")
	_, err = app.NewConfig(nil)
	assert.ErrorContains(t, err, "TASK_UI_SESSION_TTL")

	t.Setenv("TASK_UI_SESSION_TTL", "")
	_, err = app.NewConfig([]string{"--nope"})
	assert.Error(t, err)

	t.Setenv("TASK_UI_ARTIFACT_HISTORY", "-1")
	_, err = app.NewConfig(nil)
	assert.ErrorContains(t, err, "TASK_UI_ARTIFACT_HISTORY")
	t.Setenv("TASK_UI_ARTIFACT_HISTORY", "")

	t.Setenv("TASK_UI_NOW", "yesterday")
	_, err = app.NewConfig(nil)
	assert.ErrorContains(t, err, "TASK_UI_NOW")
	t.Setenv("TASK_UI_NOW", "")
}

// The data and key folders default to /data and /keys where they exist,
// else ./data and ./keys; the workspace folder to <data>/workspaces,
// made when missing. The variables of the older layout stop it.
func TestNewConfigAt_Defaults(t *testing.T) {
	for _, name := range []string{"TASK_UI_DATA", "TASK_UI_SSH_DIR", "TASK_UI_WORKSPACE", "TASK_UI_DB", "TASK_UI_ROOT", "TASK_UI_HTPASSWD"} {
		t.Setenv(name, "")
		assert.NoError(t, os.Unsetenv(name))
	}
	t.Chdir(t.TempDir())
	bare := t.TempDir()
	cfg, err := app.NewConfigAt(bare, nil)
	assert.NoError(t, err)
	assert.Equal(t, "data", cfg.DataDir)
	assert.Equal(t, "keys", cfg.SSHDir)
	assert.Equal(t, filepath.Join("data", "workspaces"), cfg.Root)
	assert.Equal(t, cfg.WorkspacesDir(), cfg.Root)
	assert.DirExists(t, cfg.Root)
	assert.Equal(t, filepath.Join("data", ".htpasswd"), cfg.Htpasswd)

	image := t.TempDir()
	assert.NoError(t, os.Mkdir(filepath.Join(image, "data"), 0o755))
	assert.NoError(t, os.Mkdir(filepath.Join(image, "keys"), 0o700))
	cfg, err = app.NewConfigAt(image, nil)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(image, "data"), cfg.DataDir)
	assert.Equal(t, filepath.Join(image, "keys"), cfg.SSHDir)
	assert.Equal(t, filepath.Join(image, "data", "workspaces"), cfg.Root)
	assert.Equal(t, filepath.Join(image, "data", ".htpasswd"), cfg.Htpasswd)
	assert.Equal(t, filepath.Join(image, "data", "task-ui.db"), cfg.DB)

	t.Setenv("TASK_UI_HOST_DATA", "/mnt/apps/task/data")
	cfg, err = app.NewConfigAt(image, nil)
	assert.NoError(t, err)
	assert.Equal(t, "/mnt/apps/task/data", cfg.HostData)

	t.Setenv("TASK_UI_ROOT", "/app")
	_, err = app.NewConfigAt(image, nil)
	assert.ErrorContains(t, err, "TASK_UI_ROOT is not read: the folders live in TASK_UI_WORKSPACE")
	assert.NoError(t, os.Unsetenv("TASK_UI_ROOT"))
	t.Setenv("TASK_UI_HTPASSWD", "/app/.htpasswd")
	_, err = app.NewConfigAt(image, nil)
	assert.ErrorContains(t, err, "the password file is <data>/.htpasswd")
}

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	cfg := &app.Config{Root: t.TempDir(), DataDir: "d", SessionTTL: time.Hour}
	assert.NoError(t, cfg.Validate())
	assert.Equal(t, filepath.Join("d", "task-ui.db"), cfg.DB)
	assert.Equal(t, filepath.Join("d", ".htpasswd"), cfg.Htpasswd)
	assert.Nil(t, cfg.Import)
	assert.True(t, time.Since(cfg.Clock()()) < time.Minute)

	assert.ErrorContains(t, (&app.Config{}).Validate(), "data directory")
	assert.ErrorContains(t, (&app.Config{Root: t.TempDir()}).Validate(), "positive")
}
