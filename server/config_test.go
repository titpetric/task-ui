package server

import (
	"path/filepath"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	root := testRoot(t)
	t.Setenv("TASK_UI_ROOT", root)
	t.Setenv("TASK_UI_HISTORY", "false")
	t.Setenv("TASK_UI_SESSION_TTL", "1h")
	cfg, err := NewConfig([]string{"-l", "127.0.0.1:9999", "--history-enable"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Listen != "127.0.0.1:9999" || !cfg.History || cfg.SessionTTL != time.Hour {
		t.Fatalf("flags over env: %+v", cfg)
	}
	if cfg.ConfigFile != filepath.Join(root, "task-ui.yml") {
		t.Fatalf("config file not found in root: %q", cfg.ConfigFile)
	}
	t.Setenv("TASK_UI_SESSION_TTL", "soon")
	if _, err := NewConfig(nil); err == nil {
		t.Fatal("bad ttl accepted")
	}
}

func TestConfig_Validate(t *testing.T) {
	cfg := &Config{Root: t.TempDir(), Columns: -1, Rows: 0}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.Columns != 120 || cfg.Rows != 30 || cfg.SessionTTL != 12*time.Hour || cfg.ConfigFile != "" {
		t.Fatalf("defaults: %+v", cfg)
	}
	cfg.EnvPath = "/data/env"
	if cfg.SecretsPath() != "/data/env/secrets" || cfg.FilesPath() != "/data/env/files" {
		t.Fatalf("env paths: %q %q", cfg.SecretsPath(), cfg.FilesPath())
	}

	// Single file mode serves the file's folder alone.
	root := testRoot(t)
	single := &Config{File: filepath.Join(root, "echo", "atkins.yml")}
	folders := single.discovery(&FileConfig{}).List()
	if len(folders) != 1 || folders[0].ID != "echo" || folders[0].Job("hello") == nil {
		t.Fatalf("single file: %+v", folders)
	}
}
