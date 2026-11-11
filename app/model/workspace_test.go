package model_test

import (
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestWorkspace_Dir(t *testing.T) {
	t.Parallel()
	root := filepath.Join("srv", "task")
	assert.Equal(t, root, (&model.Workspace{Folder: model.Default}).Dir(root))
	assert.Equal(t, filepath.Join(root, "malta"), (&model.Workspace{Folder: "malta"}).Dir(root))
}
