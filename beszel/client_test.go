package beszel_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/internal/assert"
)

// hub is a Beszel hub that takes the token t1, with two systems.
func hub(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var lists atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/systems/records" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "t1" {
			http.Error(w, "{}", http.StatusUnauthorized)
			return
		}
		lists.Add(1)
		_, _ = w.Write([]byte(`{"items":[{"id":"a","name":"dev1","host":"10.10.10.11","status":"up","info":{"h":"dev1","k":"6.12","c":4,"m":"EPYC","u":3600,"cpu":12.5,"mp":61.2,"dp":40,"v":"0.12"}},{"id":"b","name":"docker1","status":"down"}]}`))
	}))
	t.Cleanup(server.Close)
	return server, &lists
}

// The client lists the systems with its token, and keeps them for its
// TTL.
func TestClient_Systems(t *testing.T) {
	t.Parallel()
	server, lists := hub(t)
	c := &beszel.Client{URL: server.URL + "/", Token: "t1", TTL: time.Hour}

	systems, err := c.Systems(context.Background())
	assert.NoError(t, err)
	assert.Len(t, systems, 2)
	dev1 := systems[0]
	assert.True(t, dev1.Up())
	assert.Equal(t, 61.2, dev1.Info.Memory)
	assert.Equal(t, time.Hour, dev1.UptimeDuration())
	assert.False(t, systems[1].Up())

	_, err = c.Systems(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int32(1), lists.Load(), "kept for the TTL")
}

// A refused token, and a hub that does not answer, are errors.
func TestClient_Token(t *testing.T) {
	t.Parallel()
	server, _ := hub(t)
	_, err := (&beszel.Client{URL: server.URL, Token: "expired"}).Systems(context.Background())
	assert.ErrorIs(t, err, beszel.ErrToken)

	_, err = (&beszel.Client{URL: "http://127.0.0.1:1", Token: "t1"}).Systems(context.Background())
	assert.Error(t, err)
}
