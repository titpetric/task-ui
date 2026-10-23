package mcp_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/mcp"
)

// create_workflow and update_workflow take a description and the steps
// with their kinds, job, workflow, prompt and summary; the final prompt
// is a step, no field of its own.
func TestWorkflowTools_StepKinds(t *testing.T) {
	t.Parallel()
	seen := 0
	for _, tool := range mcp.New(http.NotFoundHandler(), "").Tools() {
		if tool.Name != "create_workflow" && tool.Name != "update_workflow" {
			continue
		}
		seen++
		props := tool.InputSchema.Properties
		_, ok := props["description"]
		assert.True(t, ok, tool.Name)
		_, ok = props["prompt"]
		assert.False(t, ok, tool.Name)
		_, ok = props["prompt_folder"]
		assert.False(t, ok, tool.Name)
		steps := props["steps"]
		assert.NotNil(t, steps.Items, tool.Name)
		assert.Equal(t, []string{"job", "workflow", "prompt", "summary"}, steps.Items.Properties["kind"].Enum)
	}
	assert.Equal(t, 2, seen)
}
