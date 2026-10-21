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
	assert.Len(t, group.SharedSecrets(), 1)
	assert.Len(t, group.Projects(), 0)

	group.Secrets = append(group.Secrets, model.Secret{Scope: "acme/web", Name: "B"}, model.Secret{Scope: "acme/api", Name: "A"}, model.Secret{Scope: "acme/web", Name: "C"})
	projects := group.Projects()
	assert.Len(t, projects, 2)
	assert.Equal(t, "acme/api", projects[0].Folder)
	assert.Equal(t, "acme/web", projects[1].Folder)
	assert.Len(t, projects[1].Secrets, 2)
}
