package server

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadFileConfig(t *testing.T) {
	root := testRoot(t)
	cfg, err := LoadFileConfig(filepath.Join(root, "task-ui.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Users) != 3 || cfg.Users.Find("ops") == nil || len(cfg.Projects) != 5 {
		t.Fatalf("loaded: %+v", cfg)
	}
	bad := filepath.Join(t.TempDir(), "bad.yml")
	mkfile(t, bad, "users:\n  - name: x\n    password: plain\n")
	if _, err := LoadFileConfig(bad); err == nil {
		t.Fatal("plain password accepted")
	}
	mkfile(t, bad, "users: [")
	if _, err := LoadFileConfig(bad); err == nil {
		t.Fatal("broken yaml accepted")
	}
	if _, err := LoadFileConfig(filepath.Join(t.TempDir(), "missing.yml")); err == nil {
		t.Fatal("missing file accepted")
	}
}

func TestConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "task-ui.yml")
	file := &configFile{path: path}
	if len(file.Get().Users) != 0 {
		t.Fatal("users without a file")
	}

	mkfile(t, path, "title: one\n")
	if file.Get().Title != "one" {
		t.Fatal("file not loaded")
	}

	// A change is picked up by its modification time.
	mkfile(t, path, "title: two\n")
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatal(err)
	}
	if file.Get().Title != "two" {
		t.Fatal("change not reloaded")
	}

	// A file that stops parsing keeps the previous configuration.
	mkfile(t, path, "title: [")
	later := future.Add(2 * time.Second)
	if err := os.Chtimes(path, later, later); err != nil {
		t.Fatal(err)
	}
	if file.Get().Title != "two" {
		t.Fatal("broken file replaced the configuration")
	}

	none := &configFile{}
	if none.Get() == nil {
		t.Fatal("no path, no config")
	}
}
