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

// The containers come from the containers collection, or on a hub
// without one from the newest container_stats record of each host.
func TestClient_Containers(t *testing.T) {
	t.Parallel()
	for _, collection := range []bool{true, false} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/api/collections/containers/records" && collection:
				_, _ = w.Write([]byte(`{"items":[{"system":"a","name":"db","image":"postgres:17","status":"Up 2 days","cpu":4.5,"memory":512,"net":2048}]}`))
			case r.URL.Path == "/api/collections/container_stats/records":
				_, _ = w.Write([]byte(`{"items":[{"system":"a","created":"2026-10-23 10:01:00.000Z","stats":[{"n":"db","c":4.5,"m":512,"b":[1024,1024]}]},{"system":"a","created":"2026-10-23 10:00:00.000Z","stats":[{"n":"old","c":1,"m":1}]}]}`))
			default:
				http.NotFound(w, r)
			}
		}))
		t.Cleanup(server.Close)
		containers, err := (&beszel.Client{URL: server.URL, TTL: time.Minute}).Containers(context.Background())
		assert.NoError(t, err)
		assert.Len(t, containers["a"], 1)
		db := containers["a"][0]
		assert.Equal(t, beszel.Container{Name: "db", CPU: 4.5, Memory: 512, Net: 2048, Image: db.Image, Status: db.Status}, db)
		if collection {
			assert.Equal(t, "postgres:17", db.Image)
		}
	}
}
