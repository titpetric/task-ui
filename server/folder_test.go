package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestFolderPage(t *testing.T) {
	ts, _ := testServer(t)
	c := login(t, ts, "ops")

	code, body := get(t, c, ts.URL+"/f/echo")
	if code != http.StatusOK {
		t.Fatalf("folder page: %d", code)
	}
	for _, want := range []string{"Echo", "atkins.yml", "Say hello", "Admins only", "data-run", "runs with"} {
		if !strings.Contains(body, want) {
			t.Fatalf("folder page lacks %q", want)
		}
	}
	if code, _ := get(t, c, ts.URL+"/f/missing"); code != http.StatusNotFound {
		t.Fatalf("missing folder: %d", code)
	}
}
