package spotify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/spotify"
)

func fmtJSON(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}

// A 429 is retried after the Retry-After it names, then answered.
func TestClient_RateLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newFake(t)
	f.limited = 2
	var slept []time.Duration
	c := &spotify.Client{API: f.URL + "/v1", Sleep: func(_ context.Context, d time.Duration) error { slept = append(slept, d); return nil }}
	me, err := c.Me(ctx, f.access)
	assert.NoError(t, err)
	assert.Equal(t, "u1", me.ID)
	assert.Equal(t, []time.Duration{time.Second, time.Second}, slept)

	f.limited = 10
	_, err = c.Me(ctx, f.access)
	assert.ErrorIs(t, err, spotify.ErrRateLimited)

	// A wait longer than the cap is not slept through.
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	c = &spotify.Client{API: srv.URL}
	_, err = c.Me(ctx, "t")
	assert.ErrorIs(t, err, spotify.ErrRateLimited)
	assert.Equal(t, int32(1), calls.Load())
}

// Every request has a timeout.
func TestClient_Timeout(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()
	c := &spotify.Client{API: srv.URL, HTTP: &http.Client{Timeout: 50 * time.Millisecond}}
	_, err := c.Me(context.Background(), "t")
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), srv.URL, "the address stays out of the error")
}

func TestClient_Errors(t *testing.T) {
	t.Parallel()
	f := newFake(t)
	c := &spotify.Client{API: f.URL + "/v1"}
	_, err := c.Me(context.Background(), "wrong")
	assert.ErrorIs(t, err, spotify.ErrUnauthorized)
	_, err = c.Playlists(context.Background(), f.access)
	assert.NoError(t, err)
	err = c.Volume(context.Background(), f.access, 101)
	assert.Error(t, err)
}

func TestLimiter(t *testing.T) {
	t.Parallel()
	now := time.Now()
	l := &spotify.Limiter{Max: 2, Window: time.Minute, Now: func() time.Time { return now }}
	assert.True(t, l.Allow("a"))
	assert.True(t, l.Allow("a"))
	assert.False(t, l.Allow("a"))
	assert.True(t, l.Allow("b"))
	now = now.Add(time.Minute)
	assert.True(t, l.Allow("a"))
}
