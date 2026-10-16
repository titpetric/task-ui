package folder

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestConfig_Apply(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "Makefile"), "build:\n\tgo build\n")
	f := Load("tools", dir)
	Config{Title: "Tools", Description: "The tools", Allow: []string{"ops"}, Command: "gmake", Jobs: map[string][]string{"build": {"admin"}}}.Apply(f)
	if f.Title != "Tools" || f.Description != "The tools" || !slices.Equal(f.Allow, []string{"ops"}) {
		t.Fatalf("overrides: %+v", f)
	}
	if !slices.Equal(f.Sources[0].Command, []string{"gmake"}) || !slices.Equal(f.Job("build").Command, []string{"gmake", "build"}) {
		t.Fatalf("command override: %v %v", f.Sources[0].Command, f.Job("build").Command)
	}
	if !slices.Equal(f.Job("build").Allow, []string{"admin"}) {
		t.Fatalf("job allow: %+v", f.Job("build"))
	}

	write(t, filepath.Join(dir, "package.json"), `{"scripts": {"build": "vite build"}}`)
	f = Load("tools", dir)
	Config{Command: "gmake", Commands: map[string]string{"npm": "pnpm run"}}.Apply(f)
	if !slices.Equal(f.Job("make:build").Command, []string{"make", "build"}) {
		t.Fatalf("command applies to a single source only: %v", f.Job("make:build").Command)
	}
	if !slices.Equal(f.Job("npm:build").Command, []string{"pnpm", "run", "build"}) {
		t.Fatalf("per kind command: %v", f.Job("npm:build").Command)
	}
}
