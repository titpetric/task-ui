package model_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestSchedule_Due(t *testing.T) {
	t.Parallel()
	now := time.Now()
	past := now.Add(-time.Minute)
	future := now.Add(time.Minute)
	assert.False(t, (*model.Schedule)(nil).Due(now))
	assert.False(t, (&model.Schedule{Enabled: 1}).Due(now), "no next run")
	assert.False(t, (&model.Schedule{Enabled: 0, NextRunAt: &past}).Due(now), "disabled")
	assert.False(t, (&model.Schedule{Enabled: 1, NextRunAt: &future}).Due(now))
	assert.True(t, (&model.Schedule{Enabled: 1, NextRunAt: &past}).Due(now))
	assert.True(t, (&model.Schedule{Enabled: 1, NextRunAt: &now}).Due(now))
}

func TestSchedule_IsEnabled(t *testing.T) {
	t.Parallel()
	assert.False(t, (*model.Schedule)(nil).IsEnabled())
	assert.False(t, (&model.Schedule{}).IsEnabled())
	assert.True(t, (&model.Schedule{Enabled: 1}).IsEnabled())
}
