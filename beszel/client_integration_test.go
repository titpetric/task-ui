//go:build integration

package beszel_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/internal/assert"
)

// hub answers the paths the client reads, with the systems page and the
// auth answer of the call.
func hub(t *testing.T, systems string, auth func(http.ResponseWriter)) *beszel.Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/collections/users/auth-refresh":
			auth(w)
		case "/api/collections/systems/records":
			_, _ = w.Write([]byte(systems))
		default:
			_, _ = w.Write([]byte(`{"items":[]}`))
		}
	}))
	t.Cleanup(server.Close)
	return &beszel.Client{URL: server.URL, Token: "t", TTL: time.Minute, HTTP: server.Client()}
}

// ok is a hub that accepts the token and renews it.
func ok(w http.ResponseWriter) {
	_, _ = w.Write([]byte(`{"token":"fresh"}`))
}

// Systems reads the page, and an empty page is the hub's answer only while
// the hub still accepts the token.
func TestClient_Systems(t *testing.T) {
	t.Parallel()
	client := hub(t, `{"items":[{"id":"a","name":"web","host":"web.example","status":"up"}]}`, ok)
	systems, err := client.Systems(context.Background())
	assert.NoError(t, err)
	assert.Len(t, systems, 1)
	assert.Equal(t, "web", systems[0].Name)

	empty := hub(t, `{"items":[]}`, ok)
	systems, err = empty.Systems(context.Background())
	assert.NoError(t, err)
	assert.Len(t, systems, 0)

	dead := hub(t, `{"items":[]}`, func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401}`))
	})
	_, err = dead.Systems(context.Background())
	assert.ErrorIs(t, err, beszel.ErrToken)
}

// A refused token is answered out of the cache for TTL: the hub is asked
// twice, the list and the auth probe, and not once per page view.
func TestClient_SystemsCachesTheRefusal(t *testing.T) {
	t.Parallel()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/collections/users/auth-refresh" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	t.Cleanup(server.Close)
	client := &beszel.Client{URL: server.URL, Token: "t", TTL: time.Minute, HTTP: server.Client()}

	for range 3 {
		_, err := client.Systems(context.Background())
		assert.ErrorIs(t, err, beszel.ErrToken)
	}
	assert.Equal(t, 2, calls, "the list and the probe, once")
}

// Alive returns the token the hub answers with, and ErrToken for one it
// refuses.
func TestClient_Alive(t *testing.T) {
	t.Parallel()
	token, err := hub(t, `{"items":[]}`, ok).Alive(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "fresh", token)

	_, err = hub(t, `{"items":[]}`, func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusForbidden)
	}).Alive(context.Background())
	assert.ErrorIs(t, err, beszel.ErrToken)
}
