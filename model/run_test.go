package model_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestRun_Key(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "infra/make:build", (&model.Run{Folder: "infra", Job: "make:build"}).Key())
}

func TestRun_Finished(t *testing.T) {
	t.Parallel()
	assert.False(t, (*model.Run)(nil).Finished())
	assert.False(t, (&model.Run{State: model.StateRunning}).Finished())
	assert.True(t, (&model.Run{State: model.StateOK}).Finished())
}

func TestRun_Elapsed(t *testing.T) {
	t.Parallel()
	start := time.Now().Add(-time.Minute)
	end := start.Add(10 * time.Second)
	assert.Equal(t, time.Duration(0), (*model.Run)(nil).Elapsed())
	assert.Equal(t, time.Duration(0), (&model.Run{}).Elapsed())
	assert.Equal(t, 10*time.Second, (&model.Run{StartedAt: &start, FinishedAt: &end}).Elapsed())
	assert.GreaterOrEqual(t, (&model.Run{StartedAt: &start}).Elapsed(), time.Minute)
}
