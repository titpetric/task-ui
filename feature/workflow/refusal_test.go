package workflow_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/workflow"
)

// A refusal names the step, and the run-as user when the refusal is theirs.
func TestRefusal(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", workflow.Refusal{}.String())
	assert.Equal(t, "you don't have permission to run infra / plan", workflow.Refusal{Step: "infra / plan"}.String())
	assert.Equal(t, "ops may not run infra / plan", workflow.Refusal{Step: "infra / plan", User: "ops"}.String())
}
