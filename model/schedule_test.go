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

func TestSchedule_ArgList(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"df -h /"}, (&model.Schedule{Args: `["df -h /"]`}).ArgList())
	assert.Empty(t, (&model.Schedule{}).ArgList())
}

func TestSchedule_Triages(t *testing.T) {
	t.Parallel()
	assert.False(t, (*model.Schedule)(nil).Triages())
	assert.False(t, (&model.Schedule{}).Triages())
	assert.True(t, (&model.Schedule{Triage: 1}).Triages())
}

func TestSchedule_IsWorkflow(t *testing.T) {
	t.Parallel()
	assert.False(t, (*model.Schedule)(nil).IsWorkflow())
	assert.False(t, (&model.Schedule{Folder: "infra", Job: "plan"}).IsWorkflow())
	assert.True(t, (&model.Schedule{WorkflowID: "w1"}).IsWorkflow())
}
