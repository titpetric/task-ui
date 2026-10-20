package access_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/access"
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

func TestBearerToken(t *testing.T) {
	t.Parallel()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(t, "", access.BearerToken(r))
	r.Header.Set("Authorization", "Basic abc")
	assert.Equal(t, "", access.BearerToken(r))
	r.Header.Set("Authorization", "bearer tui_abc ")
	assert.Equal(t, "tui_abc", access.BearerToken(r))
}
