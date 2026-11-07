package access_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// Attach keeps nobody out: a request with no session reaches /login with no user on it.
// Require would redirect that same request to /login.
func TestGuardAttachKeepsNobodyOut(t *testing.T) {
	t.Parallel()
	guard := &access.Guard{LoginPath: "/login"}
	var drawn bool
	var reader *model.User
	page := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		drawn = true
		reader = access.UserFrom(r.Context())
	})

	w := httptest.NewRecorder()
	guard.Attach(page).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/login", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, drawn)
	assert.Nil(t, reader)
}

func TestIsHTTPS(t *testing.T) {
	t.Parallel()
	assert.True(t, access.IsHTTPS(httptest.NewRequest(http.MethodGet, "https://x/", nil)))
	assert.False(t, access.IsHTTPS(httptest.NewRequest(http.MethodGet, "http://x/", nil)))
	r := httptest.NewRequest(http.MethodGet, "http://x/", nil)
	r.Header.Set("X-Forwarded-Proto", "HTTPS")
	assert.True(t, access.IsHTTPS(r))
}
