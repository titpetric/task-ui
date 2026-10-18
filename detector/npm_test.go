package detector

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestNpmDetector(t *testing.T) {
	dir := t.TempDir()
	if (npmDetector{}).Detect(dir) != nil {
		t.Fatal("no package.json, yet a source")
	}
	write(t, filepath.Join(dir, "package.json"), `{
  "name": "site", "description": "The marketing site",
  "scripts": {"build": "vite build", "prebuild": "rimraf dist", "dev": "vite", "postinstall": "husky"}
}`)
	source := (npmDetector{}).Detect(dir)
	if source.Name != "site" || source.Description != "The marketing site" || !slices.Equal(source.Command, []string{"npm", "run"}) {
		t.Fatalf("npm source: %+v", source)
	}
	names := []string{}
	for _, target := range source.Targets {
		names = append(names, target.Name)
	}
	if want := []string{"build", "dev", "postinstall"}; !slices.Equal(names, want) {
		t.Fatalf("npm scripts: got %v, want %v", names, want)
	}
	if source.Targets[0].Description != "vite build" {
		t.Fatalf("npm desc: %+v", source.Targets[0])
	}
	write(t, filepath.Join(dir, "pnpm-lock.yaml"), "")
	if source := (npmDetector{}).Detect(dir); !slices.Equal(source.Command, []string{"pnpm", "run"}) {
		t.Fatalf("lock file manager: %v", source.Command)
	}
	write(t, filepath.Join(dir, "package.json"), `{"scripts": [`)
	if source := (npmDetector{}).Detect(dir); source == nil || source.Error == "" {
		t.Fatalf("broken package.json: %+v", source)
	}
}
