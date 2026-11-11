package access_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestNewToken(t *testing.T) {
	t.Parallel()
	token, hash, err := access.NewToken()
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(token, access.TokenPrefix))
	assert.Len(t, token, len(access.TokenPrefix)+64)
	assert.Equal(t, access.HashToken(token), hash)
	assert.NotEqual(t, token, hash)

	other, _, _ := access.NewToken()
	assert.NotEqual(t, token, other)
}

// TokenName takes the label a reader typed, falls back to the default
// one, and counts up past the names the user already holds, because a
// name is unique per user and kind.
func TestTokenName(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "mcp", access.TokenName("", nil))
	assert.Equal(t, "laptop", access.TokenName(" laptop ", nil))
	assert.Equal(t, "mcp-2", access.TokenName("", []string{"mcp"}))
	assert.Equal(t, "mcp-4", access.TokenName("", []string{"mcp", "mcp-2", "mcp-3"}))
	assert.Equal(t, "mcp-3", access.TokenName("", []string{"mcp", "mcp-2", "mcp-4"}))
	assert.Equal(t, "laptop-2", access.TokenName("laptop", []string{"laptop"}))
}

func TestBearerToken(t *testing.T) {
	t.Parallel()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(t, "", access.BearerToken(r))
	r.Header.Set("Authorization", "Basic abc")
	assert.Equal(t, "", access.BearerToken(r))
	r.Header.Set("Authorization", "bearer tui_abc ")
	assert.Equal(t, "tui_abc", access.BearerToken(r))
}
