package detector

import (
	"path/filepath"
	"testing"
)

func TestComposerDetector(t *testing.T) {
	dir := t.TempDir()
	if (composerDetector{}).Detect(dir) != nil {
		t.Fatal("no composer.json, yet a source")
	}
	write(t, filepath.Join(dir, "composer.json"), `{
  "name": "acme/app",
  "scripts": {"test": "phpunit", "lint": ["phpcs", "phpstan"], "post-install-cmd": "echo no"},
  "scripts-descriptions": {"test": "Run the test suite"}
}`)
	source := (composerDetector{}).Detect(dir)
	if source.Name != "acme/app" || len(source.Targets) != 2 || source.Kind != "composer" {
		t.Fatalf("composer: %+v", source)
	}
	if lint := source.Targets[0]; lint.Name != "lint" || lint.Description != "phpcs" || len(lint.Steps) != 2 {
		t.Fatalf("composer lint: %+v", lint)
	}
	if test := source.Targets[1]; test.Description != "Run the test suite" {
		t.Fatalf("composer test: %+v", test)
	}
	write(t, filepath.Join(dir, "composer.json"), `{"scripts": [`)
	if source := (composerDetector{}).Detect(dir); source == nil || source.Error == "" {
		t.Fatalf("broken composer.json: %+v", source)
	}
}
