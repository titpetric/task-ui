package server

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestSecretsPage(t *testing.T) {
	ts, _ := testServer(t)
	admin := login(t, ts, "admin")
	ops := login(t, ts, "ops")

	if code, _ := post(t, ops, ts.URL+"/settings/secrets/save", url.Values{"scope": {"all"}, "key": {"GREETING"}, "value": {"nope"}}); code != http.StatusForbidden {
		t.Fatalf("runner saved a secret: %d", code)
	}

	for _, form := range []url.Values{
		{"scope": {"all"}, "key": {"GREETING"}, "value": {"hello from all"}},
		{"scope": {"printenv"}, "key": {"GREETING"}, "value": {"hello from the folder"}},
	} {
		code, location := post(t, admin, ts.URL+"/settings/secrets/save", form)
		if code != http.StatusSeeOther || !strings.Contains(location, "notice=") {
			t.Fatalf("save secret: %d %s", code, location)
		}
	}
	if _, location := post(t, admin, ts.URL+"/settings/secrets/save", url.Values{"scope": {"all"}, "key": {"bad key"}, "value": {"x"}}); !strings.Contains(location, "error=") {
		t.Fatal("invalid key accepted")
	}
	if _, location := post(t, admin, ts.URL+"/settings/secrets/save", url.Values{"scope": {"nowhere"}, "key": {"X"}, "value": {"x"}}); !strings.Contains(location, "error=") {
		t.Fatal("unknown folder accepted")
	}

	code, body := get(t, admin, ts.URL+"/settings/secrets")
	if code != http.StatusOK || strings.Contains(body, "hello from") || !strings.Contains(body, "GREETING") {
		t.Fatalf("secrets page shows values: %d", code)
	}
	code, body = get(t, admin, ts.URL+"/settings/secrets/edit?scope=printenv&key=GREETING")
	if code != http.StatusOK || !strings.Contains(body, "Edit secret") || !strings.Contains(body, "hello from the folder") {
		t.Fatalf("edit form: %d", code)
	}
	code, body = get(t, admin, ts.URL+"/settings/secrets/new?scope=printenv")
	if code != http.StatusOK || !strings.Contains(body, "Add secret") {
		t.Fatalf("add form: %d", code)
	}

	if _, location := post(t, admin, ts.URL+"/settings/secrets/delete", url.Values{"scope": {"printenv"}, "key": {"GREETING"}}); !strings.Contains(location, "notice=") {
		t.Fatalf("delete: %s", location)
	}
	if _, location := post(t, admin, ts.URL+"/settings/secrets/delete", url.Values{"scope": {"printenv"}, "key": {"GREETING"}}); !strings.Contains(location, "error=") {
		t.Fatal("deleted twice")
	}
}
