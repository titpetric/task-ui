package mcp_test

import (
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/mcp"
)

func TestNew(t *testing.T) {
	t.Parallel()
	s := mcp.New(http.NotFoundHandler(), "v1")
	assert.Equal(t, "v1", s.Version)
	names := []string{}
	for _, tool := range s.Tools() {
		names = append(names, tool.Name)
		assert.Equal(t, "object", tool.InputSchema.Type)
	}
	assert.Equal(t, []string{"list_jobs", "list_workflows", "run_job", "run_workflow", "list_runs", "get_run", "list_notes", "note_comment", "list_workflow_runs", "post_note", "ask_prompt", "list_artifacts", "read_artifact", "write_artifact", "get_workflow", "create_workflow", "update_workflow", "whoami"}, names)
}
