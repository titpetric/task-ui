package access_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

// A run token signs in as its user until its revoke, and marks the
// request as a run's with its thread; tokens of other runs stay.
func TestRunTokens(t *testing.T) {
	t.Parallel()
	tokens := access.NewRunTokens()
	token, revoke, err := tokens.Issue("u1", "n1")
	assert.NoError(t, err)
	other, _, err := tokens.Issue("u2", "")
	assert.NoError(t, err)
	id, ok := tokens.User(token)
	assert.True(t, ok)
	assert.Equal(t, "u1", id)
	assert.Equal(t, "n1", tokens.Thread(token))
	assert.Equal(t, "", tokens.Thread(other))
	revoke()
	_, ok = tokens.User(token)
	assert.False(t, ok)
	assert.Equal(t, "", tokens.Thread(token))
	id, _ = tokens.User(other)
	assert.Equal(t, "u2", id)
	_, ok = (*access.RunTokens)(nil).User(other)
	assert.False(t, ok)
}

func TestGuard_RunToken(t *testing.T) {
	t.Parallel()
	guard, ops := newGuard(t)
	guard.RunTokens = access.NewRunTokens()
	token, revoke, err := guard.RunTokens.Issue(ops.ID, "n1")
	assert.NoError(t, err)
	bearer := func() *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		return r
	}
	var fromRun bool
	var thread string
	who := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fromRun, thread = access.FromRun(r.Context()), access.RunThread(r.Context())
		_, _ = w.Write([]byte(access.UserFrom(r.Context()).Name))
	})
	w := httptest.NewRecorder()
	guard.Require(who).ServeHTTP(w, bearer())
	assert.Equal(t, "ops", w.Body.String())
	assert.True(t, fromRun)
	assert.Equal(t, "n1", thread)

	revoke()
	w = httptest.NewRecorder()
	guard.Require(who).ServeHTTP(w, bearer())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
