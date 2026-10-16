package server

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestEnvFilePage(t *testing.T) {
	ts, svc := testServer(t)
	admin := login(t, ts, "admin")

	code, location := post(t, admin, ts.URL+"/f/printenv/env/.env", url.Values{"content": {"EDITED=yes\n"}})
	if code != http.StatusSeeOther || !strings.Contains(location, "notice=") {
		t.Fatalf("save env file: %d %s", code, location)
	}
	if code, _ := post(t, admin, ts.URL+"/f/printenv/env/other.env", url.Values{"content": {"X=1"}}); code != http.StatusNotFound {
		t.Fatalf("undeclared env file accepted: %d", code)
	}
	if code, _ := get(t, admin, ts.URL+"/f/printenv/env/other.env"); code != http.StatusNotFound {
		t.Fatalf("undeclared env file page: %d", code)
	}

	code, body := get(t, admin, ts.URL+"/f/printenv/env/.env")
	if code != http.StatusOK || !strings.Contains(body, "EDITED=yes") || !strings.Contains(body, "managed") {
		t.Fatalf("env file page: %d", code)
	}
	code, body = get(t, admin, ts.URL+"/f/printenv")
	if code != http.StatusOK || !strings.Contains(body, "managed") || !strings.Contains(body, "EDITED") {
		t.Fatalf("folder page env files: %d", code)
	}

	file, err := svc.env.File(svc.folders().Find("printenv"), ".env")
	if err != nil || !file.Managed {
		t.Fatalf("stored file: %+v %v", file, err)
	}
}
