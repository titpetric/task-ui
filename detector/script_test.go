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
