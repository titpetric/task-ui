package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestAuthKind(t *testing.T) {
	t.Parallel()
	kinds := []model.AuthKind{model.AuthPassword, model.AuthBackup}
	assert.Equal(t, []string{"password", "backup"}, kinds)
}
