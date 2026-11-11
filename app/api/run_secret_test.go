package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

// fakeRunSecrets hold one run, of token "run-token", whose folder has
// the secret DEMO.
type fakeRunSecrets struct{}

func (fakeRunSecrets) Secret(_ context.Context, token, name string) (string, error) {
	if token != "run-token" {
		return "", runner.ErrRunToken
	}
	if name != "DEMO" {
		return "", runner.ErrNoSecret
	}
	return "s3cret value", nil
}

// runSecretCall sends GET or another method for a secret, with a token.
func runSecretCall(t *testing.T, method, token, name string) (int, map[string]string) {
	t.Helper()
	req := httptest.NewRequest(method, "/api/run/secrets/"+name, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	(&api.RunSecret{Source: fakeRunSecrets{}}).ServeHTTP(rec, req)
	out := map[string]string{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

// The run's own token gets a secret of its folder, and nothing else
// gets anything.
func TestRunSecret(t *testing.T) {
	t.Parallel()
	status, body := runSecretCall(t, http.MethodGet, "run-token", "DEMO")
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "s3cret value", body["value"])

	status, body = runSecretCall(t, http.MethodGet, "run-token", "OTHER")
	assert.Equal(t, http.StatusNotFound, status)
	assert.Contains(t, body["error"], "OTHER")

	status, body = runSecretCall(t, http.MethodGet, "a-user-token", "DEMO")
	assert.Equal(t, http.StatusUnauthorized, status, "a token of no running run")
	assert.Equal(t, "", body["value"])

	status, _ = runSecretCall(t, http.MethodGet, "", "DEMO")
	assert.Equal(t, http.StatusUnauthorized, status, "no token")

	status, _ = runSecretCall(t, http.MethodPost, "run-token", "DEMO")
	assert.Equal(t, http.StatusMethodNotAllowed, status)
}
