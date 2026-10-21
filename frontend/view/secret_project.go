package view

import (
	"github.com/titpetric/task-ui/model"
)

// SecretProject is a project on the secrets page, with the secrets
// scoped to it alone.
type SecretProject struct {
	Folder  string
	Secrets []model.Secret
}
