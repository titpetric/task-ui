package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
)

func TestKeyWorkspace(t *testing.T) {
	t.Parallel()
	row := view.KeyWorkspace{
		Workspace: model.Workspace{Folder: "acme"}, Projects: []string{"acme/api"}, Secrets: 2,
		SSH: []keystore.Key{{Name: "workspace_acme_deploy_ed25519"}, {Name: "workspace_acme_ssh_rsa4096"}},
	}
	assert.Nil(t, row.Encrypt)
	assert.Equal(t, []string{"workspace_acme_deploy_ed25519", "workspace_acme_ssh_rsa4096"}, row.Names())
	assert.Equal(t, []string{}, view.KeyWorkspace{}.Names())
}
