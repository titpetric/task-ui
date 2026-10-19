package detector

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestScriptDetector(t *testing.T) {
	dir := t.TempDir()
	if (scriptDetector{}).Detect(dir) != nil {
		t.Fatal("no scripts folder, yet a source")
	}
	write(t, filepath.Join(dir, "scripts", "deploy.sh"), "#!/usr/bin/env bash\n# shellcheck disable=SC2086\n# Deploy the current release\nset -e\n")
	write(t, filepath.Join(dir, "scripts", "tool"), "#!/usr/bin/env python3\n\"\"\"not a comment\"\"\"\n")
	write(t, filepath.Join(dir, "scripts", "README.md"), "# docs\n")
	write(t, filepath.Join(dir, "scripts", ".hidden.sh"), "")
	if err := os.Chmod(filepath.Join(dir, "scripts", "tool"), 0o755); err != nil {
		t.Fatal(err)
	}

	source := (scriptDetector{}).Detect(dir)
	if source == nil || len(source.Targets) != 2 || source.Kind != "script" {
		t.Fatalf("scripts: %+v", source)
	}
	deploy, tool := source.Targets[0], source.Targets[1]
	if deploy.Name != "deploy" || deploy.Description != "Deploy the current release" || !slices.Equal(deploy.Args, []string{"sh", "scripts/deploy.sh"}) {
		t.Fatalf("deploy: %+v", deploy)
	}
	if tool.Name != "tool" || tool.Description != "" || !slices.Equal(tool.Args, []string{"./scripts/tool"}) {
		t.Fatalf("tool: %+v", tool)
	}

	write(t, filepath.Join(dir, "scripts", "README.md"), "")
	os.Remove(filepath.Join(dir, "scripts", "deploy.sh"))
	os.Remove(filepath.Join(dir, "scripts", "tool"))
	if (scriptDetector{}).Detect(dir) != nil {
		t.Fatal("a scripts folder with nothing runnable is a source")
	}
}

// A scripts folder that runs its scripts itself, from a pipeline file
// or a Makefile of its own, is a folder of its own: its scripts are not
// listed again on the folder above.
func TestScriptDetector_OwnedScripts(t *testing.T) {
	for name, file := range map[string]string{
		"atkins.yml":   "jobs:\n  update: ./update.sh\n",
		"Taskfile.yml": "tasks:\n  update:\n    cmds: [./update.sh]\n",
		"Makefile":     "update:\n\t./update.sh\n",
	} {
		dir := t.TempDir()
		write(t, filepath.Join(dir, "scripts", "update.sh"), "#!/usr/bin/env bash\n# Update\n")
		write(t, filepath.Join(dir, "scripts", name), file)
		if source := (scriptDetector{}).Detect(dir); source != nil {
			t.Fatalf("%s owns the scripts, yet they are listed above it: %+v", name, source)
		}
	}
}
