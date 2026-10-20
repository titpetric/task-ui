package mcp_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/mcp"
)

// A tool is listed with its name, what it does and its arguments.
func TestTool(t *testing.T) {
	t.Parallel()
	for _, tool := range mcp.New(http.NotFoundHandler(), "").Tools() {
		var listed mcp.Tool = tool
		assert.NotEqual(t, "", listed.Description, listed.Name)
		data, err := json.Marshal(listed)
		assert.NoError(t, err)
		assert.Contains(t, string(data), `"inputSchema":{"type":"object"`)
	}
}
