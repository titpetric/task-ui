package server

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestAgo(t *testing.T) {
	now := time.Now()
	for d, want := range map[time.Duration]string{
		10 * time.Second: "just now",
		5 * time.Minute:  "5m ago",
		3 * time.Hour:    "3h ago",
		72 * time.Hour:   "3d ago",
	} {
		if got := ago(now.Add(-d)); got != want {
			t.Errorf("%v: got %q, want %q", d, got, want)
		}
	}
}

func TestDurationText(t *testing.T) {
	for d, want := range map[time.Duration]string{
		500 * time.Millisecond:  "500ms",
		3700 * time.Millisecond: "3.7s",
		90 * time.Second:        "1m30s",
		2 * time.Hour:           "2h00m",
	} {
		if got := durationText(d); got != want {
			t.Errorf("%v: got %q, want %q", d, got, want)
		}
	}
}

func TestRender(t *testing.T) {
	ts, svc := testServer(t)
	c := login(t, ts, "guest")
	code, body := get(t, c, ts.URL+"/f/nope")
	if code != http.StatusNotFound || !strings.Contains(body, "No such folder") {
		t.Fatalf("not found page: %d", code)
	}
	resp, err := c.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.Header.Get("X-Frame-Options") != "DENY" {
		t.Fatal("frame options missing")
	}

	// A template that cannot be found is a server error, not a panic.
	w := &recorder{header: http.Header{}}
	svc.render(w, resp.Request, "missing.tpl", &pageData{})
	if w.status != http.StatusInternalServerError {
		t.Fatalf("render without a template: %d", w.status)
	}
}

// recorder is the little of http.ResponseWriter render needs.
type recorder struct {
	header http.Header
	status int
}

func (r *recorder) Header() http.Header         { return r.header }
func (r *recorder) Write(p []byte) (int, error) { return len(p), nil }
func (r *recorder) WriteHeader(status int)      { r.status = status }
