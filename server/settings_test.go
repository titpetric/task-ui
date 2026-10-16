package server

import (
	"net/http"
	"strings"
	"testing"
)

func TestSettingsPage(t *testing.T) {
	ts, _ := testServer(t)
	admin := login(t, ts, "admin")
	code, body := get(t, admin, ts.URL+"/settings")
	if code != http.StatusOK {
		t.Fatalf("settings: %d", code)
	}
	for _, want := range []string{"Projects", "SSH keys", "Secrets", "4 folders"} {
		if !strings.Contains(body, want) {
			t.Fatalf("settings lacks %q", want)
		}
	}
}
