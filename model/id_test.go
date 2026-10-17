package model_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestNewID(t *testing.T) {
	t.Parallel()
	a, b := model.NewID(), model.NewID()
	assert.True(t, model.ValidID(a), a)
	assert.True(t, model.ValidID(b), b)
	assert.NotEqual(t, a, b)
}

func TestNewIDAt(t *testing.T) {
	t.Parallel()
	earlier := model.NewIDAt(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	later := model.NewIDAt(time.Date(2026, 1, 2, 3, 4, 6, 0, time.UTC))
	assert.Less(t, earlier, later, "ids sort by time")
	assert.Len(t, earlier, 26)
}

func TestValidID(t *testing.T) {
	t.Parallel()
	assert.False(t, model.ValidID(""))
	assert.False(t, model.ValidID("../x"))
	assert.False(t, model.ValidID("ABCDEFGHJKMNPQRSTVWXYZ0123"))
	assert.True(t, model.ValidID("0123456789abcdefghjkmnpqrs"))
}
