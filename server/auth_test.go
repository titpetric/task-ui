package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSafeBack(t *testing.T) {
	for in, want := range map[string]string{
		"":              "/",
		"/f/echo":       "/f/echo",
		"//evil":        "/",
		"/\\evil":       "/",
		"http://x":      "/",
		"/settings?x=1": "/settings?x=1",
	} {
		if got := safeBack(in); got != want {
			t.Errorf("safeBack(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSameHost(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "http://app.example/x", nil)
	r.Header.Set("X-Forwarded-Host", "public.example, other.example")
	if !sameHost(r, "app.example") || !sameHost(r, "public.example") || !sameHost(r, "other.example") {
		t.Fatal("same host refused")
	}
	if sameHost(r, "evil.example") || sameHost(r, "") {
		t.Fatal("other host accepted")
	}
	if !isHTTPS(httptest.NewRequest(http.MethodGet, "https://x/", nil)) {
		t.Fatal("https not seen")
	}
	r.Header.Set("X-Forwarded-Proto", "https")
	if !isHTTPS(r) {
		t.Fatal("forwarded https not seen")
	}
}

func TestRequireUser(t *testing.T) {
	ts, _ := testServer(t)
	c := client()

	resp, err := c.Get(ts.URL + "/f/echo/hello")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther || !strings.HasPrefix(resp.Header.Get("Location"), "/login?back=") {
		t.Fatalf("anonymous page: %d %s", resp.StatusCode, resp.Header.Get("Location"))
	}
	if code, _ := get(t, c, ts.URL+"/api/runs"); code != http.StatusUnauthorized {
		t.Fatalf("api without session: %d", code)
	}

	// A forged cookie is nobody.
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/runs", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "a.b.c"})
	resp, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("forged cookie: %d", resp.StatusCode)
	}
}

func TestRequireAdmin(t *testing.T) {
	ts, _ := testServer(t)
	ops := login(t, ts, "ops")
	for _, path := range []string{"/settings", "/settings/keys", "/settings/secrets", "/settings/projects", "/f/printenv/env/.env"} {
		if code, _ := get(t, ops, ts.URL+path); code != http.StatusForbidden {
			t.Fatalf("runner %s: %d", path, code)
		}
	}
}

func TestSameOrigin(t *testing.T) {
	ts, _ := testServer(t)
	c := login(t, ts, "admin")

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/settings/keys/generate", strings.NewReader("type=ed25519"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "http://evil.example")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("cross origin post: %d", resp.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodPost, ts.URL+"/logout", nil)
	req.Header.Set("Referer", ts.URL+"/")
	resp, err = c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("same origin referer refused: %d", resp.StatusCode)
	}
}
