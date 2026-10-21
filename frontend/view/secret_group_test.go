package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestSecretGroup(t *testing.T) {
	t.Parallel()
	group := view.SecretGroup{Workspace: "acme", Title: "Acme", Secrets: []model.Secret{{Scope: "workspace:acme", Name: "TOKEN"}}, Folders: []string{"acme/api"}}
	assert.Equal(t, "workspace:acme", group.Shared())
	assert.Len(t, group.Secrets, 1)
	assert.Equal(t, []string{"acme/api"}, group.Folders)
}
