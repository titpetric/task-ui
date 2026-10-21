package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestIsInternal(t *testing.T) {
	t.Parallel()
	assert.True(t, model.IsInternal(model.Internal))
	assert.True(t, model.IsInternal("internal/tools"))
	assert.False(t, model.IsInternal("internals"))
	assert.False(t, model.IsInternal("deploy/internal"))
	assert.False(t, model.IsInternal(""))
}

func TestIsReserved(t *testing.T) {
	t.Parallel()
	assert.True(t, model.IsReserved(model.Internal))
	assert.True(t, model.IsReserved("internal/tools"))
	assert.True(t, model.IsReserved(model.Default))
	assert.True(t, model.IsReserved("_/web"))
	assert.False(t, model.IsReserved("_web"))
	assert.False(t, model.IsReserved("web/_"))
	assert.False(t, model.IsReserved("internals"))
	assert.False(t, model.IsReserved(""))
}
