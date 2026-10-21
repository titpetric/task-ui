package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestSecretProject(t *testing.T) {
	t.Parallel()
	project := view.SecretProject{Folder: "acme/api", Secrets: []model.Secret{{Scope: "acme/api", Name: "TOKEN"}}}
	assert.Equal(t, "acme/api", project.Folder)
	assert.Len(t, project.Secrets, 1)
}
