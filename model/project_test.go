package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestProject_AllowList(t *testing.T) {
	t.Parallel()
	assert.Len(t, (*model.Project)(nil).AllowList(), 0)
	assert.Equal(t, []string{"ops"}, (&model.Project{Allow: `["ops"]`}).AllowList())
}

func TestProject_SourceList(t *testing.T) {
	t.Parallel()
	assert.Len(t, (*model.Project)(nil).SourceList(), 0)
	assert.Equal(t, []string{"make"}, (&model.Project{Sources: `["make"]`}).SourceList())
}

func TestProject_CommandMap(t *testing.T) {
	t.Parallel()
	assert.Len(t, (*model.Project)(nil).CommandMap(), 0)
	assert.Equal(t, map[string]string{"make": "gmake"}, (&model.Project{Commands: `{"make":"gmake"}`}).CommandMap())
}

func TestProject_JobAllowMap(t *testing.T) {
	t.Parallel()
	assert.Len(t, (*model.Project)(nil).JobAllowMap(), 0)
	assert.Equal(t, map[string][]string{"deploy": {"admin"}}, (&model.Project{JobAllow: `{"deploy":["admin"]}`}).JobAllowMap())
}
