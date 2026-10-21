package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/sshkey"
)

func TestKeyWorkspace(t *testing.T) {
	t.Parallel()
	row := view.KeyWorkspace{
		Workspace: model.Workspace{Folder: "acme"}, Projects: []string{"acme/api"}, Secrets: 2,
		SSH: []sshkey.Key{{Name: "id_ed25519_deploy"}, {Name: "id_rsa"}},
	}
	assert.Nil(t, row.Encrypt)
	assert.Equal(t, []string{"id_ed25519_deploy", "id_rsa"}, row.Names())
	assert.Equal(t, []string{}, view.KeyWorkspace{}.Names())
}
