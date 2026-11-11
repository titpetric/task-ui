package secret_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/secret"
)

// runSecrets hold one run, of token "run-token", whose folder has DEMO.
type runSecrets struct{}

func (runSecrets) Secret(_ context.Context, token, name string) (string, error) {
	if token != "run-token" {
		return "", runner.ErrRunToken
	}
	if name != "DEMO" {
		return "", runner.ErrNoSecret
	}
	return "s3cret", nil
}

// handlerTransport serves a request with a handler, in process.
type handlerTransport struct{ h http.Handler }

func (t handlerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	t.h.ServeHTTP(rec, r)
	return rec.Result(), nil
}

// client returns a client of the run's environment, whose requests the
// route serves in process.
func client(t *testing.T, token string) *secret.Client {
	t.Helper()
	env := map[string]string{"TASK_UI_API": "http://127.0.0.1:3000/api/", "TASK_UI_RUN_TOKEN": token}
	c, err := secret.ClientFromEnv(func(name string) string { return env[name] })
	assert.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:3000/api", c.API)
	c.HTTP = &http.Client{Transport: handlerTransport{&api.RunSecret{Source: runSecrets{}}}}
	return c
}

// A run gets a secret of its folder; a missing one, a token of no
// running run and a name with a path in it fail with a message.
func TestClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	value, err := client(t, "run-token").Get(ctx, "DEMO")
	assert.NoError(t, err)
	assert.Equal(t, "s3cret", value)

	_, err = client(t, "run-token").Get(ctx, "MISSING")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")

	_, err = client(t, "ended").Get(ctx, "DEMO")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "401")

	_, err = client(t, "run-token").Get(ctx, "../keys")
	assert.Error(t, err)

	forbidden := &secret.Client{API: "http://x/api", Token: "t", HTTP: &http.Client{Transport: handlerTransport{http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
	})}}}
	_, err = forbidden.Get(ctx, "DEMO")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")

	_, err = secret.ClientFromEnv(func(string) string { return "" })
	assert.Error(t, err, "outside a run")
}
