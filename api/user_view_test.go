package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A user's view is the row with its MCP access beside it, flat.
func TestUserView(t *testing.T) {
	t.Parallel()
	view := api.UserView{User: model.User{ID: "u1", Name: "ops", Role: model.RoleRunner}}
	raw, err := json.Marshal(view)
	assert.NoError(t, err)
	assert.Contains(t, string(raw), `"name":"ops"`)
	assert.Contains(t, string(raw), `"mcp":{"enabled":false}`)
}
