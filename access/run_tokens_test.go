package access_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

// A run token signs in as its user until its revoke, and marks the
// request as a run's; tokens of other runs stay.
func TestRunTokens(t *testing.T) {
	t.Parallel()
	tokens := access.NewRunTokens()
	token, revoke, err := tokens.Issue("u1")
	assert.NoError(t, err)
	other, _, err := tokens.Issue("u2")
	assert.NoError(t, err)
	id, ok := tokens.User(token)
	assert.True(t, ok)
	assert.Equal(t, "u1", id)
	revoke()
	_, ok = tokens.User(token)
	assert.False(t, ok)
	id, _ = tokens.User(other)
	assert.Equal(t, "u2", id)
	_, ok = (*access.RunTokens)(nil).User(other)
	assert.False(t, ok)
}

func TestGuard_RunToken(t *testing.T) {
	t.Parallel()
	guard, ops := newGuard(t)
	guard.RunTokens = access.NewRunTokens()
	token, revoke, err := guard.RunTokens.Issue(ops.ID)
	assert.NoError(t, err)
	bearer := func() *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		return r
	}
	var fromRun bool
	who := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fromRun = access.FromRun(r.Context())
		_, _ = w.Write([]byte(access.UserFrom(r.Context()).Name))
	})
	w := httptest.NewRecorder()
	guard.Require(who).ServeHTTP(w, bearer())
	assert.Equal(t, "ops", w.Body.String())
	assert.True(t, fromRun)

	revoke()
	w = httptest.NewRecorder()
	guard.Require(who).ServeHTTP(w, bearer())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
