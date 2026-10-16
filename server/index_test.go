package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestIndex(t *testing.T) {
	ts, _ := testServer(t)
	c := login(t, ts, "ops")

	ws, err := dial(t, ts, c, "/ws/run/echo/hello", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	collect(t, ws, final)
	ws.Close()

	code, body := get(t, c, ts.URL+"/")
	if code != http.StatusOK || !strings.Contains(body, "Overview") {
		t.Fatalf("overview: %d", code)
	}
	for _, want := range []string{"echo/", "sleep/", "Recent runs", "hello"} {
		if !strings.Contains(body, want) {
			t.Fatalf("overview lacks %q", want)
		}
	}
}
