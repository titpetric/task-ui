package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshkey"
)

func TestKeyFolder(t *testing.T) {
	t.Parallel()
	folder := view.KeyFolder{Project: "deploy", Dir: "/app/deploy/.ssh", Keys: []sshkey.Key{{Name: "id_rsa"}}}
	assert.Equal(t, "deploy", folder.Project)
	assert.Len(t, folder.Keys, 1)
}
