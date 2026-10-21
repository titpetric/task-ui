package detector_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/internal/assert"
)

// A file runs as itself when executable, with its shebang when not,
// and not at all without either.
func TestFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(dir, "bin"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "bin", "deploy"), []byte("#!/bin/sh\n# Ship it\n# interactive: true\necho ok\n"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "tool.py"), []byte("#!/usr/bin/env python3 -u\nprint(1)\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("just text\n"), 0o644))

	deploy := detector.File(dir, "bin/deploy")
	assert.NotNil(t, deploy)
	assert.Equal(t, "bin/deploy", deploy.Name)
	assert.Equal(t, "Ship it", deploy.Description)
	assert.True(t, deploy.Interactive)
	assert.Equal(t, []string{"./bin/deploy"}, deploy.Args)

	tool := detector.File(dir, "tool.py")
	assert.NotNil(t, tool)
	assert.Equal(t, []string{"/usr/bin/env", "python3", "-u", "tool.py"}, tool.Args)

	assert.Nil(t, detector.File(dir, "notes.txt"))
	assert.Nil(t, detector.File(dir, "bin"))
	assert.Nil(t, detector.File(dir, "missing"))
}

// A shebang is the first line's interpreter and arguments.
func TestShebang(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for name, content := range map[string]string{"a": "#!/bin/bash -e\n", "b": "#!", "c": "echo\n", "d": "#!/bin/sh"} {
		assert.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
	}
	assert.Equal(t, []string{"/bin/bash", "-e"}, detector.Shebang(filepath.Join(dir, "a")))
	assert.Nil(t, detector.Shebang(filepath.Join(dir, "b")))
	assert.Nil(t, detector.Shebang(filepath.Join(dir, "c")))
	assert.Equal(t, []string{"/bin/sh"}, detector.Shebang(filepath.Join(dir, "d")))
	assert.Nil(t, detector.Shebang(filepath.Join(dir, "missing")))
}
