package server

import (
	"bytes"
	"html"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKeysPage(t *testing.T) {
	ts, svc := testServer(t)
	admin := login(t, ts, "admin")

	code, body := get(t, admin, ts.URL+"/settings/keys")
	if code != http.StatusOK || !strings.Contains(body, "No key pair yet") {
		t.Fatalf("empty keys page: %d", code)
	}
	for _, page := range []string{"/settings/keys/generate", "/settings/keys/upload"} {
		if code, body := get(t, admin, ts.URL+page); code != http.StatusOK || !strings.Contains(body, "Cancel") {
			t.Fatalf("%s: %d", page, code)
		}
	}

	code, location := post(t, admin, ts.URL+"/settings/keys/generate", url.Values{"type": {"ed25519"}, "comment": {"ci key"}})
	if code != http.StatusSeeOther || !strings.Contains(location, "notice=") {
		t.Fatalf("generate: %d %s", code, location)
	}
	keys := svc.keys.List()
	if len(keys) != 1 || keys[0].Comment != "ci-key" {
		t.Fatalf("keys: %+v", keys)
	}
	code, body = get(t, admin, ts.URL+"/settings/keys")
	if code != http.StatusOK || !strings.Contains(html.UnescapeString(body), keys[0].Fingerprint) {
		t.Fatalf("keys page: %d", code)
	}
	if _, location := post(t, admin, ts.URL+"/settings/keys/generate", url.Values{"type": {"dsa"}}); !strings.Contains(location, "error=") {
		t.Fatal("unsupported type accepted")
	}

	if _, location := post(t, admin, ts.URL+"/settings/keys/delete", url.Values{"name": {"id_ed25519"}}); !strings.Contains(location, "notice=") {
		t.Fatalf("delete: %s", location)
	}
	if len(svc.keys.List()) != 0 {
		t.Fatal("key still there")
	}
	if _, location := post(t, admin, ts.URL+"/settings/keys/delete", url.Values{"name": {"../x"}}); !strings.Contains(location, "error=") {
		t.Fatal("bad name accepted")
	}
}

func TestKeysUpload(t *testing.T) {
	ts, svc := testServer(t)
	admin := login(t, ts, "admin")

	// Borrow a key from another store to upload.
	other := svc.keys
	other.Dir = filepath.Join(t.TempDir(), "other")
	if _, err := other.Generate("ed25519", "elsewhere"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(other.Dir, "id_ed25519"))
	svc.keys.Dir = filepath.Join(t.TempDir(), "ssh")

	if _, location := post(t, admin, ts.URL+"/settings/keys/upload", url.Values{"private_key": {string(data)}, "comment": {"laptop"}}); !strings.Contains(location, "notice=") {
		t.Fatalf("pasted upload: %s", location)
	}
	if keys := svc.keys.List(); len(keys) != 1 || keys[0].Comment != "laptop" {
		t.Fatalf("after upload: %+v", keys)
	}

	// The same through the file field, as a browser sends it.
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, _ := form.CreateFormFile("key_file", "id_ed25519")
	part.Write(data)
	form.WriteField("comment", "file")
	form.Close()
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/settings/keys/upload", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	resp, err := admin.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Location"), "notice=") {
		t.Fatalf("file upload: %s", resp.Header.Get("Location"))
	}

	if _, location := post(t, admin, ts.URL+"/settings/keys/upload", url.Values{"private_key": {" "}}); !strings.Contains(location, "error=") {
		t.Fatal("empty upload accepted")
	}
	if _, location := post(t, admin, ts.URL+"/settings/keys/upload", url.Values{"private_key": {"garbage"}}); !strings.Contains(location, "error=") {
		t.Fatal("garbage upload accepted")
	}
}
