package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestLogin(t *testing.T) {
	ts, _ := testServer(t)
	c := client()

	if code, body := get(t, c, ts.URL+"/login"); code != http.StatusOK || !strings.Contains(body, "Sign in") {
		t.Fatalf("login page: %d", code)
	}
	if code, _ := post(t, c, ts.URL+"/login", url.Values{"username": {"ops"}, "password": {"wrong"}}); code != http.StatusUnauthorized {
		t.Fatalf("wrong password: %d", code)
	}
	if code, _ := post(t, c, ts.URL+"/login", url.Values{"username": {"nobody"}, "password": {"x"}}); code != http.StatusUnauthorized {
		t.Fatalf("unknown user: %d", code)
	}

	// An off-site back target is ignored.
	if _, location := post(t, c, ts.URL+"/login", url.Values{"username": {"ops"}, "password": {"ops"}, "back": {"//evil.example/"}}); location != "/" {
		t.Fatalf("open redirect: %s", location)
	}
	code, body := get(t, c, ts.URL+"/f/echo/hello")
	if code != http.StatusOK || !strings.Contains(body, "Say hello") || !strings.Contains(body, "data-run") {
		t.Fatalf("job page after login: %d", code)
	}
}

func TestLogout(t *testing.T) {
	ts, _ := testServer(t)
	c := login(t, ts, "ops")
	if code, location := post(t, c, ts.URL+"/logout", nil); code != http.StatusSeeOther || location != "/login" {
		t.Fatalf("logout: %d %s", code, location)
	}
	if code, _ := get(t, c, ts.URL+"/api/runs"); code != http.StatusUnauthorized {
		t.Fatalf("session survived logout: %d", code)
	}
}

func TestLoginNoUsers(t *testing.T) {
	cfg := &Config{Root: t.TempDir(), SessionTTL: 1}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	svc, err := NewWithConfig(os.DirFS(".."), cfg)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(svc.Handler())
	defer ts.Close()

	c := client()
	resp, _ := c.Get(ts.URL + "/")
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("overview without users: %d", resp.StatusCode)
	}
	code, body := get(t, c, ts.URL+"/login")
	if code != http.StatusServiceUnavailable || !strings.Contains(body, "No users are configured") {
		t.Fatalf("login without users: %d", code)
	}
}

func TestLoginAutoLogin(t *testing.T) {
	ts, svc := testServer(t)
	svc.config.AutoLogin = "ops"

	c := client()
	code, body := get(t, c, ts.URL+"/f/echo")
	if code != http.StatusOK || !strings.Contains(body, "auto login") || !strings.Contains(body, "data-run") {
		t.Fatalf("auto login as ops: %d", code)
	}
	if code, _ := get(t, c, ts.URL+"/settings"); code != http.StatusForbidden {
		t.Fatalf("auto login kept the role: %d", code)
	}
	if code, location := post(t, c, ts.URL+"/login", url.Values{"back": {"/f/echo"}}); code != http.StatusSeeOther || location != "/f/echo" {
		t.Fatalf("login page under auto login: %d %s", code, location)
	}

	svc.config.AutoLogin = "nobody"
	if code, _ := get(t, c, ts.URL+"/api/runs"); code != http.StatusUnauthorized {
		t.Fatalf("auto login as an unknown user: %d", code)
	}
}
