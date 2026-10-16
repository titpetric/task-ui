package server

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestAPIRuns(t *testing.T) {
	ts, _ := testServer(t)
	ops := login(t, ts, "ops")

	if code, body := get(t, ops, ts.URL+"/api/runs"); code != http.StatusOK || strings.TrimSpace(body) != "[]" {
		t.Fatalf("no runs: %d %q", code, body)
	}

	ws, err := dial(t, ts, ops, "/ws/run/sleep/5", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	collect(t, ws, running)

	_, body := get(t, ops, ts.URL+"/api/runs")
	if !strings.Contains(body, `"key":"sleep/5"`) || !strings.Contains(body, `"user":"ops"`) {
		t.Fatalf("api runs: %s", body)
	}
}

func TestAPIStop(t *testing.T) {
	ts, svc := testServer(t)
	ops := login(t, ts, "ops")
	guest := login(t, ts, "guest")
	admin := login(t, ts, "admin")

	ws, err := dial(t, ts, ops, "/ws/run/sleep/5", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	collect(t, ws, running)

	if code, _ := post(t, guest, ts.URL+"/api/runs/sleep/5/stop", nil); code != http.StatusForbidden {
		t.Fatalf("viewer killed a run: %d", code)
	}
	if code, _ := post(t, admin, ts.URL+"/api/runs/sleep/nothing/stop", nil); code != http.StatusNotFound {
		t.Fatalf("stop of a missing job: %d", code)
	}
	code, location := post(t, admin, ts.URL+"/api/runs/sleep/5/stop", url.Values{"back": {"/f/sleep"}})
	if code != http.StatusSeeOther || location != "/f/sleep" {
		t.Fatalf("kill: %d %s", code, location)
	}
	if _, statuses := collect(t, ws, final); len(statuses) == 0 || statuses[len(statuses)-1].State != "killed" {
		t.Fatalf("after kill: %+v", statuses)
	}
	waitFor(t, func() bool {
		_, busy := svc.runs.Get("sleep/5")
		return !busy
	})

	// Stopping what is not running is fine, and JSON clients get 204.
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/runs/sleep/5/stop", nil)
	req.Header.Set("Accept", "application/json")
	resp, err := admin.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("json stop: %d", resp.StatusCode)
	}
}
