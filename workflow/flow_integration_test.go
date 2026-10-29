//go:build integration

package workflow_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// FlowByID loads a workflow with its steps, ErrNotFound for none.
func TestRunner_FlowByID(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	wf := f.create(t, "build", "test")

	flow, err := f.runner.FlowByID(ctx, wf.ID)
	assert.NoError(t, err)
	assert.Equal(t, wf.ID, flow.Workflow().ID)
	assert.Len(t, flow.Steps(), 2)
	_, err = f.runner.FlowByID(ctx, "nowhere")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

// Start starts a run of the flow as the user.
func TestFlow_Start(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	flow, err := f.runner.FlowByID(ctx, f.create(t, "build").ID)
	assert.NoError(t, err)
	run, err := flow.Start(ctx, ops, "")
	assert.NoError(t, err)
	assert.Equal(t, model.StateOK, f.wait(t, run.ID).State)
}
