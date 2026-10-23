//go:build integration

package detector

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestFromFile(t *testing.T) {
	// FromFile treats a yaml file no convention claims as an atkins
	// pipeline, and asks atkins to list it.
	stubAtkins(t, `[{"kind":"pipeline","active":true,"desc":"Jobs","cmds":[
	  {"id":"sync","name":"sync","desc":"Sync","cmd":"atkins sync","active":true}]}]`)

	dir := t.TempDir()
	write(t, filepath.Join(dir, "jobs.yml"), "jobs:\n  sync:\n    desc: Sync\n")
	source := FromFile(filepath.Join(dir, "jobs.yml"))
	if source == nil || source.Kind != "atkins" || source.Target("sync") == nil {
		t.Fatalf("unknown yaml name read as atkins: %+v", source)
	}
	if !slices.Equal(source.Command, []string{"atkins", "-f", "jobs.yml"}) {
		t.Fatalf("command: %v", source.Command)
	}

	write(t, filepath.Join(dir, "Taskfile.yml"), "tasks:\n  build:\n    desc: Build\n")
	if source := FromFile(filepath.Join(dir, "Taskfile.yml")); source == nil || source.Kind != "task" {
		t.Fatalf("Taskfile by name: %+v", source)
	}
	if FromFile(filepath.Join(dir, "README.md")) != nil {
		t.Fatal("markdown read as a source")
	}
	if FromFile(filepath.Join(dir, "missing.yml")) != nil {
		t.Fatal("missing file read as a source")
	}
	if FromFile(dir) != nil {
		t.Fatal("directory read as a source")
	}
}
