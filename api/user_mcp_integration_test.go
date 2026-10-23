//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// User/edit enables, rotates and disables anyone's MCP token; a user
// rotates or disables their own once on, and enables nothing.
// The POST that makes a token is the one answer that holds it.
func TestUserMCP(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	ops := f.users[asOps]
	path := "/users/" + ops.ID + "/mcp"

	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPost, path, nil).Status, "enabling your own takes user/edit")
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPost, "/users/"+f.users[asViewer].ID+"/mcp", nil).Status)
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPost, "/users/"+model.NewID()+"/mcp", nil).Status)
	assert.Equal(t, http.StatusNotFound, f.call(t, asAdmin, http.MethodPost, "/users/"+model.NewID()+"/mcp", nil).Status)

	enabled := f.call(t, asAdmin, http.MethodPost, path, nil)
	assert.Equal(t, http.StatusCreated, enabled.Status, string(enabled.Raw))
	token := enabled.str("token")
	assert.True(t, strings.HasPrefix(token, access.TokenPrefix))
	expires, err := time.Parse(time.RFC3339Nano, enabled.str("expires_at"))
	assert.NoError(t, err)
	assert.True(t, expires.Sub(time.Now()) > access.MCPTokenTTL-time.Minute)
	row, err := f.opts.Tokens.Find(ctx, access.HashToken(token))
	assert.NoError(t, err)
	assert.Equal(t, access.MCPToken, row.Name)

	// The view of the user tells MCP is on, and never holds the token.
	var mcp api.UserMCP
	for _, raw := range []response{f.call(t, asAdmin, http.MethodGet, "/users", nil), f.call(t, asAdmin, http.MethodPut, "/users/"+ops.ID, api.UserRequest{})} {
		assert.NotContains(t, string(raw.Raw), token)
		assert.NotContains(t, string(raw.Raw), row.Hash)
	}
	listed := f.call(t, asAdmin, http.MethodGet, "/users", nil)
	for _, u := range listed.list("users") {
		entry := u.(map[string]any)
		if entry["name"] != asOps {
			continue
		}
		data, _ := json.Marshal(entry["mcp"])
		assert.NoError(t, json.Unmarshal(data, &mcp))
	}
	assert.True(t, mcp.Enabled)
	assert.NotNil(t, mcp.ExpiresAt)
	assert.True(t, expires.Equal(*mcp.ExpiresAt))

	// Their own, once on: ops rotates it, and the old hash finds nothing.
	rotated := f.call(t, asOps, http.MethodPost, path, nil)
	assert.Equal(t, http.StatusCreated, rotated.Status)
	assert.NotEqual(t, token, rotated.str("token"))
	_, err = f.opts.Tokens.Find(ctx, access.HashToken(token))
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.Equal(t, http.StatusForbidden, f.call(t, asViewer, http.MethodDelete, path, nil).Status)
	assert.Equal(t, http.StatusNoContent, f.call(t, asOps, http.MethodDelete, path, nil).Status)
	_, err = f.opts.Tokens.Get(ctx, ops.ID, access.MCPToken)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPost, path, nil).Status, "off again, enabling takes user/edit")
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodDelete, path, nil).Status)
}
