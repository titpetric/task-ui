package server

import (
	"net/http"
	"testing"
)

func TestRoutes(t *testing.T) {
	ts, _ := testServer(t)
	c := client()
	if code, body := get(t, c, ts.URL+"/healthz"); code != http.StatusOK || body != "ok\n" {
		t.Fatalf("healthz: %d %q", code, body)
	}
	resp, err := c.Get(ts.URL + "/static/task-ui.css")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || resp.Header.Get("Cache-Control") == "" {
		t.Fatalf("static: %d %q", resp.StatusCode, resp.Header.Get("Cache-Control"))
	}
	if code, _ := get(t, c, ts.URL+"/static/missing.css"); code != http.StatusNotFound {
		t.Fatalf("missing asset: %d", code)
	}
}
