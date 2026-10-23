package beszel_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/internal/assert"
)

// monthHub has no eight hour records yet, only two hour ones: two of
// host a, one of b, and a month of a's containers.
func monthHub(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var asks atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asks.Add(1)
		filter := r.URL.Query().Get("filter")
		if !strings.Contains(filter, "type='120m'") || !strings.Contains(filter, "created>=") {
			_, _ = w.Write([]byte(`{"items":[]}`))
			return
		}
		switch r.URL.Path {
		case "/api/collections/system_stats/records":
			_, _ = w.Write([]byte(`{"items":[
				{"system":"a","created":"2026-10-01 08:00:00.000Z","stats":{"cpu":10,"m":8,"mu":4,"d":100,"du":50,"b":[100,200],"efs":{"data":{"d":2000,"du":1000}}}},
				{"system":"b","created":"2026-10-01 08:00:00.000Z","stats":{"cpu":1,"ns":1,"nr":2,"b":"odd"}},
				{"system":"a","created":"2026-10-01 10:00:00.000Z","stats":{"cpu":20}},
				{"system":"a","created":"not a time","stats":{"cpu":30}}]}`))
		case "/api/collections/container_stats/records":
			if !strings.Contains(filter, "system='a'") {
				http.Error(w, "{}", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte(`{"items":[{"system":"a","created":"2026-10-01 08:00:00.000Z","stats":[{"n":"db","c":3,"m":300,"ns":1,"nr":1}]}]}`))
		}
	}))
	t.Cleanup(server.Close)
	return server, &asks
}

// The month falls back to the two hour records, and the client keeps it.
func TestClient_History(t *testing.T) {
	t.Parallel()
	server, asks := monthHub(t)
	c := &beszel.Client{URL: server.URL}
	history, err := c.History(context.Background())
	assert.NoError(t, err)
	assert.Len(t, history["a"], 2)
	a := history["a"][0]
	assert.Equal(t, time.Date(2026, 10, 1, 8, 0, 0, 0, time.UTC), a.At)
	sent, recv := a.Stats.Net()
	assert.Equal(t, 100.0, sent)
	assert.Equal(t, 200.0, recv)
	assert.Equal(t, 2000.0, a.Stats.ExtraFs["data"].Total)
	sent, recv = history["b"][0].Stats.Net()
	assert.Equal(t, 1024*1024.0, sent, "MB/s of an older agent, the odd field ignored")
	assert.Equal(t, 2*1024*1024.0, recv)

	_, _ = c.History(context.Background())
	assert.Equal(t, int32(2), asks.Load(), "480m, then 120m, then kept")

	month, err := c.ContainerHistory(context.Background(), "a")
	assert.NoError(t, err)
	assert.Len(t, month, 1)
	assert.Equal(t, 2*1024*1024.0, month[0].Containers[0].Net)
	none, err := c.ContainerHistory(context.Background(), "a' || 1=1")
	assert.NoError(t, err)
	assert.Len(t, none, 0)
}

// A sample is a record's time and measures.
func TestSample(t *testing.T) {
	t.Parallel()
	s := beszel.Sample{Stats: beszel.Stats{CPU: 5}}
	assert.Equal(t, 5.0, s.Stats.CPU)
	cs := beszel.ContainerSample{Containers: []beszel.Container{{Name: "db"}}}
	assert.Equal(t, "db", cs.Containers[0].Name)
	assert.Equal(t, 0.0, beszel.FsStats{}.Total)
	assert.Equal(t, beszel.Pair{1, 2}, beszel.Pair{1, 2})
}
