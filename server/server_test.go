package server

import (
	"encoding/json"
	"html"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/websocket"
)

// testServer starts task-ui over a temporary root with two projects:
// "echo" runs `echo <job>`, "sleep" runs `sleep <job>`.
func testServer(t *testing.T) (*httptest.Server, *Server) {
	t.Helper()
	root := t.TempDir()

	mkfile := func(path, body string) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mkfile(filepath.Join(root, "echo", "atkins.yml"), `
name: Echo
jobs:
  hello:
    desc: "Say hello"
  locked:
    desc: "Admins only"
`)
	mkfile(filepath.Join(root, "printenv", "atkins.yml"), `
name: Printenv
env:
  include: .env
jobs:
  GREETING:
    desc: "Print the greeting"
`)
	mkfile(filepath.Join(root, "cat", "atkins.yml"), `
jobs:
  "-":
    desc: "Echo what is typed"
    interactive: true
`)
	mkfile(filepath.Join(root, "sleep", "Taskfile.yml"), `
tasks:
  "5":
    desc: "Sleep a while"
`)

	hash := func(password string) string {
		out, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
		if err != nil {
			t.Fatal(err)
		}
		return string(out)
	}
	mkfile(filepath.Join(root, "task-ui.yml"), `
users:
  - name: admin
    password: "`+hash("admin")+`"
    role: admin
  - name: ops
    password: "`+hash("ops")+`"
    role: runner
  - name: guest
    password: "`+hash("guest")+`"
    role: viewer
projects:
  - id: echo
    command: echo
    jobs:
      locked: [admin]
  - id: sleep
    command: sleep
  - id: printenv
    command: printenv
  - id: cat
    command: cat
  - id: notes
    command: cat
`)

	cfg := &Config{
		Listen:      "127.0.0.1:0",
		Root:        root,
		EnvPath:     filepath.Join(t.TempDir(), "env"),
		Discover:    true,
		History:     true,
		HistoryPath: filepath.Join(t.TempDir(), "history"),
		SSHDir:      filepath.Join(t.TempDir(), "ssh"),
		SessionTTL:  time.Hour,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}

	svc, err := NewWithConfig(os.DirFS(".."), cfg)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(svc.Handler())
	t.Cleanup(ts.Close)
	return ts, svc
}

func client(t *testing.T) *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Jar: jar,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func login(t *testing.T, ts *httptest.Server, user string) *http.Client {
	t.Helper()
	c := client(t)
	resp, err := c.PostForm(ts.URL+"/login", url.Values{"username": {user}, "password": {user}, "back": {"/f/echo"}})
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/f/echo" {
		t.Fatalf("login %s: %d %s", user, resp.StatusCode, resp.Header.Get("Location"))
	}
	return c
}

func get(t *testing.T, c *http.Client, u string) (int, string) {
	t.Helper()
	resp, err := c.Get(u)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var b strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		b.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return resp.StatusCode, b.String()
}

// dial dials the run socket as the client's user.
func dial(t *testing.T, ts *httptest.Server, c *http.Client, path, origin string) (*websocket.Conn, error) {
	t.Helper()
	u, _ := url.Parse(ts.URL)
	config, err := websocket.NewConfig("ws://"+u.Host+path, origin)
	if err != nil {
		t.Fatal(err)
	}
	for _, cookie := range c.Jar.Cookies(u) {
		config.Header.Add("Cookie", cookie.Name+"="+cookie.Value)
	}
	return websocket.DialConfig(config)
}

// collect reads frames until a final status, returning the output and the
// statuses seen.
func collect(t *testing.T, ws *websocket.Conn, until func(control) bool) (string, []control) {
	t.Helper()
	var out strings.Builder
	var statuses []control
	ws.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		var frame []byte
		if err := websocket.Message.Receive(ws, &frame); err != nil {
			return out.String(), statuses
		}
		if len(frame) > 0 && frame[0] == '{' {
			var msg control
			if json.Unmarshal(frame, &msg) == nil && msg.Type == "status" {
				statuses = append(statuses, msg)
				if until(msg) {
					return out.String(), statuses
				}
				continue
			}
		}
		out.Write(frame)
	}
}

func final(msg control) bool { return msg.State != "running" }

func TestLoginFlow(t *testing.T) {
	ts, _ := testServer(t)
	c := client(t)

	resp, err := c.Get(ts.URL + "/f/echo/hello")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther || !strings.HasPrefix(resp.Header.Get("Location"), "/login?back=") {
		t.Fatalf("anonymous: %d %s", resp.StatusCode, resp.Header.Get("Location"))
	}

	if code, _ := get(t, c, ts.URL+"/api/runs"); code != http.StatusUnauthorized {
		t.Fatalf("api without session: %d", code)
	}

	resp, _ = c.PostForm(ts.URL+"/login", url.Values{"username": {"ops"}, "password": {"wrong"}})
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong password: %d", resp.StatusCode)
	}

	// An off-site back target is ignored.
	resp, _ = c.PostForm(ts.URL+"/login", url.Values{"username": {"ops"}, "password": {"ops"}, "back": {"//evil.example/"}})
	resp.Body.Close()
	if resp.Header.Get("Location") != "/" {
		t.Fatalf("open redirect: %s", resp.Header.Get("Location"))
	}

	code, body := get(t, c, ts.URL+"/f/echo/hello")
	if code != http.StatusOK || !strings.Contains(body, "Say hello") || !strings.Contains(body, "data-run") {
		t.Fatalf("job page: %d", code)
	}
	if code, _ := get(t, c, ts.URL+"/f/nope"); code != http.StatusNotFound {
		t.Fatalf("missing project: %d", code)
	}
}

func TestViewerCannotRun(t *testing.T) {
	ts, _ := testServer(t)
	c := login(t, ts, "guest")

	_, body := get(t, c, ts.URL+"/f/echo/hello")
	if strings.Contains(body, "data-run ") || !strings.Contains(body, "not allowed to run") {
		t.Fatal("viewer page offers a run button")
	}

	ws, err := dial(t, ts, c, "/ws/run/echo/hello", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	_, statuses := collect(t, ws, final)
	if len(statuses) != 1 || statuses[0].State != "denied" {
		t.Fatalf("viewer run: %+v", statuses)
	}
}

func TestRunnerAllowList(t *testing.T) {
	ts, _ := testServer(t)
	c := login(t, ts, "ops")
	ws, err := dial(t, ts, c, "/ws/run/echo/locked", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	if _, statuses := collect(t, ws, final); len(statuses) != 1 || statuses[0].State != "denied" {
		t.Fatalf("allow list ignored: %+v", statuses)
	}
}

func TestRunRecordsHistory(t *testing.T) {
	ts, svc := testServer(t)
	c := login(t, ts, "ops")

	ws, err := dial(t, ts, c, "/ws/run/echo/hello?cols=100&rows=20", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	out, statuses := collect(t, ws, final)
	ws.Close()

	if !strings.Contains(out, "hello") {
		t.Fatalf("output: %q", out)
	}
	last := statuses[len(statuses)-1]
	if statuses[0].State != "running" || last.State != "ok" || last.Exit == nil || *last.Exit != 0 {
		t.Fatalf("statuses: %+v", statuses)
	}

	records := svc.history.List("echo", "hello")
	if len(records) != 1 || records[0].User != "ops" || records[0].State != "ok" {
		t.Fatalf("history: %+v", records)
	}
	code, cast := get(t, c, ts.URL+"/api/history/echo/hello/"+records[0].ID)
	if code != http.StatusOK || !strings.Contains(cast, "hello") {
		t.Fatalf("cast: %d %q", code, cast)
	}
	if code, _ := get(t, c, ts.URL+"/api/history/echo/hello/..%2F..%2Fetc"); code != http.StatusNotFound {
		t.Fatalf("history traversal: %d", code)
	}
}

func TestRunLockForceAndClose(t *testing.T) {
	ts, svc := testServer(t)
	alice := login(t, ts, "ops")
	bob := login(t, ts, "admin")

	first, err := dial(t, ts, alice, "/ws/run/sleep/5", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	if _, statuses := collect(t, first, func(c control) bool { return c.State == "running" }); len(statuses) == 0 {
		t.Fatal("first run did not start")
	}

	// A second run of the same job is refused.
	second, err := dial(t, ts, bob, "/ws/run/sleep/5", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, statuses := collect(t, second, final)
	second.Close()
	if len(statuses) != 1 || statuses[0].State != "busy" || statuses[0].User != "ops" {
		t.Fatalf("second run: %+v", statuses)
	}

	// The run shows up in the API.
	_, body := get(t, bob, ts.URL+"/api/runs")
	if !strings.Contains(body, `"key":"sleep/5"`) {
		t.Fatalf("api runs: %s", body)
	}

	// Run force kills the first run and takes over.
	forced, err := dial(t, ts, bob, "/ws/run/sleep/5?force=1", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if _, statuses := collect(t, forced, func(c control) bool { return c.State == "running" }); len(statuses) == 0 || statuses[len(statuses)-1].State != "running" {
		t.Fatalf("forced run: %+v", statuses)
	}
	if _, statuses := collect(t, first, final); len(statuses) == 0 || statuses[len(statuses)-1].State != "killed" {
		t.Fatalf("first run after force: %+v", statuses)
	}

	// Closing the window stops the run and frees the lock.
	forced.Close()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, busy := svc.runs.Get("sleep/5"); !busy {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("closing the socket did not stop the run")
		}
		time.Sleep(20 * time.Millisecond)
	}
	waitFor(t, func() bool {
		last, ok := svc.history.Last("sleep", "5")
		return ok && last.State == "stopped" && last.Forced
	})
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("condition not met")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestCrossOriginRefused(t *testing.T) {
	ts, _ := testServer(t)
	c := login(t, ts, "admin")

	if _, err := dial(t, ts, c, "/ws/run/echo/hello", "http://evil.example"); err == nil {
		t.Fatal("cross origin websocket accepted")
	}

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
}

func TestKeysAdminOnly(t *testing.T) {
	ts, svc := testServer(t)

	ops := login(t, ts, "ops")
	if code, _ := get(t, ops, ts.URL+"/settings/keys"); code != http.StatusForbidden {
		t.Fatalf("runner keys page: %d", code)
	}

	admin := login(t, ts, "admin")
	resp, err := admin.PostForm(ts.URL+"/settings/keys/generate", url.Values{"type": {"ed25519"}, "comment": {"ci key"}})
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther || !strings.Contains(resp.Header.Get("Location"), "notice=") {
		t.Fatalf("generate: %d %s", resp.StatusCode, resp.Header.Get("Location"))
	}
	keys := svc.keys.List()
	if len(keys) != 1 || keys[0].Comment != "ci-key" {
		t.Fatalf("keys: %+v", keys)
	}
	code, body := get(t, admin, ts.URL+"/settings/keys")
	if code != http.StatusOK || !strings.Contains(html.UnescapeString(body), keys[0].Fingerprint) {
		t.Fatalf("keys page: %d", code)
	}
}

func TestSessions(t *testing.T) {
	s := newSessions("secret", time.Hour)
	token := s.Issue("ops")
	if name, err := s.Verify(token); err != nil || name != "ops" {
		t.Fatalf("verify: %q %v", name, err)
	}
	if _, err := newSessions("other", time.Hour).Verify(token); err == nil {
		t.Fatal("token verified with another secret")
	}
	s.now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	if _, err := s.Verify(token); err == nil {
		t.Fatal("expired token verified")
	}
}

func TestNoUsersNoWayIn(t *testing.T) {
	cfg := &Config{Root: t.TempDir(), SessionTTL: time.Hour}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	svc, err := NewWithConfig(os.DirFS(".."), cfg)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(svc.Handler())
	defer ts.Close()

	c := client(t)
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

func TestAutoLogin(t *testing.T) {
	ts, svc := testServer(t)
	svc.config.AutoLogin = "ops"

	c := client(t)
	code, body := get(t, c, ts.URL+"/f/echo")
	if code != http.StatusOK || !strings.Contains(body, "auto login") || !strings.Contains(body, "data-run") {
		t.Fatalf("auto login as ops: %d", code)
	}
	if code, _ := get(t, c, ts.URL+"/settings"); code != http.StatusForbidden {
		t.Fatalf("auto login kept the role: %d", code)
	}

	svc.config.AutoLogin = "nobody"
	if code, _ := get(t, c, ts.URL+"/api/runs"); code != http.StatusUnauthorized {
		t.Fatalf("auto login as an unknown user: %d", code)
	}
}

// gitRepo makes a bare repository holding a pipeline file and a note, and
// returns its file URL and a working clone to push changes from.
func gitRepo(t *testing.T) (string, string) {
	t.Helper()
	git := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t", "-c", "init.defaultBranch=main"}, args...)...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	work := filepath.Join(t.TempDir(), "work")
	os.MkdirAll(work, 0o755)
	git(work, "init", "-q")
	os.WriteFile(filepath.Join(work, "atkins.yml"), []byte("name: Notes\njobs:\n  note.txt:\n    desc: \"Print the note\"\n"), 0o644)
	os.WriteFile(filepath.Join(work, "note.txt"), []byte("v1\n"), 0o644)
	git(work, "add", ".")
	git(work, "commit", "-q", "-m", "first")
	bare := filepath.Join(t.TempDir(), "notes.git")
	git(filepath.Dir(bare), "clone", "-q", "--bare", work, bare)
	git(work, "remote", "add", "origin", bare)
	return "file://" + bare, work
}

func TestProjectsCloneAndPull(t *testing.T) {
	ts, svc := testServer(t)
	admin := login(t, ts, "admin")
	ops := login(t, ts, "ops")
	repo, work := gitRepo(t)

	if code, _ := get(t, ops, ts.URL+"/settings/projects"); code != http.StatusForbidden {
		t.Fatalf("runner projects page: %d", code)
	}
	resp, _ := ops.PostForm(ts.URL+"/settings/projects/clone", url.Values{"url": {repo}})
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("runner cloned: %d", resp.StatusCode)
	}

	// A clone that cannot work reports git's output and keeps nothing.
	resp, _ = admin.PostForm(ts.URL+"/settings/projects/clone", url.Values{"url": {"file:///nonexistent/repo.git"}, "id": {"broken"}})
	resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Location"), "error=") {
		t.Fatalf("bad clone: %s", resp.Header.Get("Location"))
	}
	if _, err := os.Stat(filepath.Join(svc.config.Root, "broken")); err == nil {
		t.Fatal("failed clone left a folder")
	}
	resp, _ = admin.PostForm(ts.URL+"/settings/projects/clone", url.Values{"url": {"-oops"}})
	resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Location"), "error=") {
		t.Fatal("option-like URL accepted")
	}

	// The folder name comes from the URL.
	resp, _ = admin.PostForm(ts.URL+"/settings/projects/clone", url.Values{"url": {repo}})
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther || !strings.Contains(resp.Header.Get("Location"), "notice=") {
		t.Fatalf("clone: %d %s", resp.StatusCode, resp.Header.Get("Location"))
	}
	code, body := get(t, admin, ts.URL+"/settings/projects")
	if code != http.StatusOK || !strings.Contains(body, "notes") || !strings.Contains(body, "atkins.yml") {
		t.Fatalf("projects page: %d", code)
	}
	code, body = get(t, admin, ts.URL+"/f/notes")
	if code != http.StatusOK || !strings.Contains(body, "Source") || !strings.Contains(body, "first") {
		t.Fatalf("folder page source: %d", code)
	}

	// The run pulls first, so a change pushed upstream is what runs.
	ws, err := dial(t, ts, ops, "/ws/run/notes/note.txt", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	out, statuses := collect(t, ws, final)
	ws.Close()
	if !strings.Contains(out, "git pull") || !strings.Contains(out, "v1") || statuses[len(statuses)-1].State != "ok" {
		t.Fatalf("first run: %q %+v", out, statuses)
	}

	os.WriteFile(filepath.Join(work, "note.txt"), []byte("v2\n"), 0o644)
	for _, args := range [][]string{{"commit", "-q", "-am", "second"}, {"push", "-q", "origin", "HEAD:main"}} {
		cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
		cmd.Dir = work
		if o, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, o)
		}
	}
	ws, _ = dial(t, ts, ops, "/ws/run/notes/note.txt", ts.URL)
	out, _ = collect(t, ws, final)
	ws.Close()
	if !strings.Contains(out, "v2") {
		t.Fatalf("run did not pull: %q", out)
	}

	// Removing takes the checkout away; a local folder cannot be removed.
	resp, _ = admin.PostForm(ts.URL+"/settings/projects/delete", url.Values{"id": {"echo"}})
	resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Location"), "error=") {
		t.Fatal("removed a local folder")
	}
	resp, _ = admin.PostForm(ts.URL+"/settings/projects/delete", url.Values{"id": {"notes"}})
	resp.Body.Close()
	if _, err := os.Stat(filepath.Join(svc.config.Root, "notes")); err == nil {
		t.Fatal("checkout still there")
	}
}

func TestSecretsReachRuns(t *testing.T) {
	ts, svc := testServer(t)
	admin := login(t, ts, "admin")
	ops := login(t, ts, "ops")

	// Settings are the admin's.
	for _, path := range []string{"/settings", "/settings/secrets", "/settings/secrets/new", "/f/printenv/env/.env"} {
		if code, _ := get(t, ops, ts.URL+path); code != http.StatusForbidden {
			t.Fatalf("runner %s: %d", path, code)
		}
	}
	resp, _ := ops.PostForm(ts.URL+"/settings/secrets/save", url.Values{"scope": {"all"}, "key": {"GREETING"}, "value": {"nope"}})
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("runner saved a secret: %d", resp.StatusCode)
	}

	// A shared secret, a folder secret that wins, and a managed env file.
	for _, form := range []url.Values{
		{"scope": {"all"}, "key": {"GREETING"}, "value": {"hello from all"}},
		{"scope": {"printenv"}, "key": {"GREETING"}, "value": {"hello from the folder"}},
	} {
		resp, err := admin.PostForm(ts.URL+"/settings/secrets/save", form)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusSeeOther || !strings.Contains(resp.Header.Get("Location"), "notice=") {
			t.Fatalf("save secret: %d %s", resp.StatusCode, resp.Header.Get("Location"))
		}
	}
	resp, _ = admin.PostForm(ts.URL+"/settings/secrets/save", url.Values{"scope": {"all"}, "key": {"bad key"}, "value": {"x"}})
	resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Location"), "error=") {
		t.Fatal("invalid key accepted")
	}

	resp, _ = admin.PostForm(ts.URL+"/f/printenv/env/.env", url.Values{"content": {"EDITED=yes\n"}})
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther || !strings.Contains(resp.Header.Get("Location"), "notice=") {
		t.Fatalf("save env file: %d %s", resp.StatusCode, resp.Header.Get("Location"))
	}
	resp, _ = admin.PostForm(ts.URL+"/f/printenv/env/other.env", url.Values{"content": {"X=1"}})
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("undeclared env file accepted: %d", resp.StatusCode)
	}

	code, body := get(t, admin, ts.URL+"/settings/secrets")
	if code != http.StatusOK || strings.Contains(body, "hello from") || !strings.Contains(body, "GREETING") {
		t.Fatalf("secrets page shows values: %d", code)
	}
	code, body = get(t, admin, ts.URL+"/f/printenv")
	if code != http.StatusOK || !strings.Contains(body, "managed") || !strings.Contains(body, "EDITED") {
		t.Fatalf("folder page env files: %d", code)
	}

	// The run sees the folder secret, and the env file was generated.
	ws, err := dial(t, ts, ops, "/ws/run/printenv/GREETING", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	out, statuses := collect(t, ws, final)
	ws.Close()
	if !strings.Contains(out, "hello from the folder") || statuses[len(statuses)-1].State != "ok" {
		t.Fatalf("run output %q, statuses %+v", out, statuses)
	}
	project := svc.projects().Find("printenv")
	data, err := os.ReadFile(filepath.Join(project.Dir, ".env"))
	if err != nil || !strings.Contains(string(data), "EDITED=yes") || !strings.Contains(string(data), "GREETING=\"hello from the folder\"") {
		t.Fatalf("generated env file: %q %v", data, err)
	}

	// Deleting the folder secret falls back to the shared one.
	resp, _ = admin.PostForm(ts.URL+"/settings/secrets/delete", url.Values{"scope": {"printenv"}, "key": {"GREETING"}})
	resp.Body.Close()
	ws, _ = dial(t, ts, ops, "/ws/run/printenv/GREETING", ts.URL)
	out, _ = collect(t, ws, final)
	ws.Close()
	if !strings.Contains(out, "hello from all") {
		t.Fatalf("shared secret not applied: %q", out)
	}
}

func TestInteractiveInput(t *testing.T) {
	ts, _ := testServer(t)
	c := login(t, ts, "ops")

	ws, err := dial(t, ts, c, "/ws/run/cat/-", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	if _, statuses := collect(t, ws, func(m control) bool { return m.State == "running" }); len(statuses) == 0 {
		t.Fatal("run did not start")
	}

	// Typed input is a control frame; cat echoes it through the pty.
	if err := websocket.Message.Send(ws, `{"type":"input","message":"hi there\n"}`); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	for !strings.Contains(out.String(), "hi there") {
		var frame []byte
		if err := websocket.Message.Receive(ws, &frame); err != nil {
			t.Fatalf("waiting for the echo: %v, got %q", err, out.String())
		}
		if len(frame) > 0 && frame[0] != '{' {
			out.Write(frame)
		}
	}

	// Stop ends the run as stopped.
	if err := websocket.Message.Send(ws, `{"type":"stop"}`); err != nil {
		t.Fatal(err)
	}
	_, statuses := collect(t, ws, final)
	if len(statuses) == 0 || statuses[len(statuses)-1].State != "stopped" {
		t.Fatalf("after stop: %+v", statuses)
	}
}
