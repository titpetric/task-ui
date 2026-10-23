package app_test

import (
	"testing"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
)

// The runs reach the MCP server over the loopback of its address.
func TestMCPURL(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "http://127.0.0.1:3000/api/mcp", app.MCPURL("[::]:3000"))
	assert.Equal(t, "http://127.0.0.1:3000/api/mcp", app.MCPURL(":3000"))
	assert.Equal(t, "http://127.0.0.1:3000/api/mcp", app.MCPURL("0.0.0.0:3000"))
	assert.Equal(t, "http://10.0.0.5:8080/api/mcp", app.MCPURL("10.0.0.5:8080"))
	assert.Equal(t, "", app.MCPURL(""))
}
