package api_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A user's view is the row with its MCP access beside it, flat: the
// tokens as a list, so a client reads one shape whether the user holds
// none, one or several.
func TestUserView(t *testing.T) {
	t.Parallel()
	view := api.UserView{User: model.User{ID: "u1", Name: "ops", Role: model.RoleRunner}}
	raw, err := json.Marshal(view)
	assert.NoError(t, err)
	assert.Contains(t, string(raw), `"name":"ops"`)
	assert.Contains(t, string(raw), `"mcp":{"enabled":false,"tokens":null}`)

	stamp := time.Date(2026, time.October, 22, 12, 0, 0, 0, time.UTC)
	view.MCP = api.UserMCP{Enabled: true, Tokens: []api.UserMCPToken{{ID: "t1", Name: "mcp", CreatedAt: &stamp}}}
	raw, err = json.Marshal(view)
	assert.NoError(t, err)
	assert.Contains(t, string(raw), `"tokens":[{"id":"t1","name":"mcp","created_at":"2026-10-22T12:00:00Z"}]`)
	assert.NotContains(t, string(raw), "hash", "nothing a request signs with crosses the boundary")
}
