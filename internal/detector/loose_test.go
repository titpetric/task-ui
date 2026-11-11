package detector

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// A scripts folder with a pipeline of its own lists, beside it, the
// scripts its pipeline does not run: by target name or in a step, a
// script is the pipeline's. Any other folder lists none.
func TestLoose(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "scripts")
	write(t, filepath.Join(dir, "atkins.yml"), "tasks:\n  update:\n    desc: \"Update\"\n    cmds:\n      - ./update.sh\n  weather:\n    desc: \"Weather\"\n    cmds:\n      - phpscript weather.php\n")
	write(t, filepath.Join(dir, "bash.sh"), "#!/bin/bash\n# A shell in deploy/scripts\n# interactive: true\nexec bash\n")
	write(t, filepath.Join(dir, "update.sh"), "#!/bin/sh\necho update\n")
	write(t, filepath.Join(dir, "weather.sh"), "#!/bin/sh\necho named like a target\n")
	write(t, filepath.Join(dir, "weather.php"), "<?php echo 1;\n")
	write(t, filepath.Join(dir, "plain.sh"), "#!/bin/sh\necho not executable\n")
	for _, name := range []string{"bash.sh", "update.sh", "weather.sh"} {
		if err := os.Chmod(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	pipeline := pipelineDetector{kind: "atkins", names: atkinsNames}.Detect(dir)
	if pipeline == nil {
		t.Fatal("no pipeline")
	}

	loose := Loose(dir, pipeline)
	if loose == nil || !loose.Aside || loose.Kind != "script" {
		t.Fatalf("loose: %+v", loose)
	}
	names := []string{}
	for _, target := range loose.Targets {
		names = append(names, target.Name)
	}
	if !slices.Equal(names, []string{"bash", "plain"}) {
		t.Fatalf("loose scripts: %v", names)
	}
	bash := loose.Target("bash")
	if !bash.Interactive || bash.Description != "A shell in deploy/scripts" || !slices.Equal(bash.Args, []string{"./bash.sh"}) {
		t.Fatalf("bash: %+v", bash)
	}
	if plain := loose.Target("plain"); plain.Interactive || !slices.Equal(plain.Args, []string{"sh", "plain.sh"}) {
		t.Fatalf("plain: %+v", plain)
	}

	other := t.TempDir()
	write(t, filepath.Join(other, "atkins.yml"), "tasks:\n  a:\n    desc: \"A\"\n    cmds:\n      - echo a\n")
	write(t, filepath.Join(other, "run.sh"), "#!/bin/sh\n")
	if Loose(other, pipelineDetector{kind: "atkins", names: atkinsNames}.Detect(other)) != nil {
		t.Fatal("a folder not named scripts lists loose scripts")
	}
	if Loose(dir, nil) != nil {
		t.Fatal("without a pipeline there is nothing to be beside")
	}
}
