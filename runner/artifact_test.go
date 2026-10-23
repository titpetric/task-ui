package runner_test

import (
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

func TestContentType(t *testing.T) {
	t.Parallel()
	assert.Contains(t, runner.ContentType("out.txt"), "text/plain")
	assert.Equal(t, "application/json", runner.ContentType("data.JSON"))
	assert.Equal(t, "application/octet-stream", runner.ContentType("binary"))
}

func TestManager_ArtifactPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	m := runner.New(runner.Options{ArtifactRoot: root})
	artifact := &model.Artifact{Path: "run1/sub/out.txt"}
	assert.Equal(t, filepath.Join(root, "run1", "sub", "out.txt"), m.ArtifactPath(artifact))
}
