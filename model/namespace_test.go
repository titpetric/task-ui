package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestCanonical(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "internal", model.Canonical("_"))
	assert.Equal(t, "internal/web", model.Canonical("_/web"))
	assert.Equal(t, "internal", model.Canonical("internal"))
	assert.Equal(t, "_web", model.Canonical("_web"))
	assert.Equal(t, "web/_", model.Canonical("web/_"))
	assert.Equal(t, "", model.Canonical(""))
}

func TestIsReserved(t *testing.T) {
	t.Parallel()
	assert.True(t, model.IsReserved(model.Default))
	assert.True(t, model.IsReserved("internal/tools"))
	assert.True(t, model.IsReserved(model.Legacy))
	assert.True(t, model.IsReserved("_/web"))
	assert.False(t, model.IsReserved("_web"))
	assert.False(t, model.IsReserved("web/_"))
	assert.False(t, model.IsReserved("internals"))
	assert.False(t, model.IsReserved(""))
}
