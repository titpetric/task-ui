package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestJobPage(t *testing.T) {
	ts, _ := testServer(t)
	guest := login(t, ts, "guest")

	code, body := get(t, guest, ts.URL+"/f/echo/hello")
	if code != http.StatusOK || !strings.Contains(body, "Say hello") || !strings.Contains(body, "echo hello") {
		t.Fatalf("job page: %d", code)
	}
	if strings.Contains(body, "data-run ") || !strings.Contains(body, "not allowed to run") {
		t.Fatal("viewer page offers a run button")
	}
	if code, _ := get(t, guest, ts.URL+"/f/echo/nothing"); code != http.StatusNotFound {
		t.Fatalf("missing job: %d", code)
	}
}

func TestHistoryFile(t *testing.T) {
	ts, svc := testServer(t)
	ops := login(t, ts, "ops")

	ws, err := dial(t, ts, ops, "/ws/run/echo/hello", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	collect(t, ws, final)
	ws.Close()

	records := svc.history.List("echo", "hello")
	if len(records) != 1 {
		t.Fatalf("history: %+v", records)
	}
	code, cast := get(t, ops, ts.URL+"/api/history/echo/hello/"+records[0].ID)
	if code != http.StatusOK || !strings.Contains(cast, "hello") {
		t.Fatalf("cast: %d %q", code, cast)
	}
	if code, _ := get(t, ops, ts.URL+"/api/history/echo/hello/..%2F..%2Fetc"); code != http.StatusNotFound {
		t.Fatalf("history traversal: %d", code)
	}
	if code, _ := get(t, ops, ts.URL+"/api/history/echo/nothing/"+records[0].ID); code != http.StatusNotFound {
		t.Fatalf("history of a missing job: %d", code)
	}
	code, body := get(t, ops, ts.URL+"/f/echo/hello")
	if code != http.StatusOK || !strings.Contains(body, records[0].ID) {
		t.Fatal("job page does not list the run")
	}
}
