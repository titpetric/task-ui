package access_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestIsHTTPS(t *testing.T) {
	t.Parallel()
	assert.True(t, access.IsHTTPS(httptest.NewRequest(http.MethodGet, "https://x/", nil)))
	assert.False(t, access.IsHTTPS(httptest.NewRequest(http.MethodGet, "http://x/", nil)))
	r := httptest.NewRequest(http.MethodGet, "http://x/", nil)
	r.Header.Set("X-Forwarded-Proto", "HTTPS")
	assert.True(t, access.IsHTTPS(r))
}
