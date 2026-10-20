package mcp_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/mcp"
)

// A tool's arguments are an object whose required properties it has.
func TestSchema(t *testing.T) {
	t.Parallel()
	for _, tool := range mcp.New(http.NotFoundHandler(), "").Tools() {
		var schema mcp.Schema = tool.InputSchema
		assert.Equal(t, "object", schema.Type, tool.Name)
		assert.NotNil(t, schema.Properties, tool.Name)
		for _, name := range schema.Required {
			_, ok := schema.Properties[name]
			assert.True(t, ok, tool.Name+" requires "+name)
		}
	}
}

// A property describes itself; an enum lists its values.
func TestProperty(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(mcp.Property{Type: "string", Description: "A state.", Enum: []string{"ok"}})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"type":"string","description":"A state.","enum":["ok"]}`, string(data))
}
