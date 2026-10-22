package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestProfile_Activity(t *testing.T) {
	t.Parallel()
	p := &view.Profile{}
	assert.False(t, p.Activity())
	p.Prompts = []model.Note{{Body: "why"}}
	assert.True(t, p.Activity())
}
