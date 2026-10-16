package server

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
)

func mkfile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// testRoot writes the folders and users the tests run against: "echo"
// runs `echo <job>`, "sleep" runs `sleep <job>`, "cat" reads its input,
// "printenv" prints a variable, and three users, one per role, with their
// name as their password.
func testRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mkfile(t, filepath.Join(root, "echo", "atkins.yml"), "name: Echo\njobs:\n  hello:\n    desc: \"Say hello\"\n  locked:\n    desc: \"Admins only\"\n")
	mkfile(t, filepath.Join(root, "printenv", "atkins.yml"), "name: Printenv\nenv:\n  include: .env\njobs:\n  GREETING:\n    desc: \"Print the greeting\"\n")
	mkfile(t, filepath.Join(root, "cat", "atkins.yml"), "jobs:\n  \"-\":\n    desc: \"Echo what is typed\"\n    interactive: true\n")
	mkfile(t, filepath.Join(root, "sleep", "Taskfile.yml"), "tasks:\n  \"5\":\n    desc: \"Sleep a while\"\n")

	users := "users:\n"
	for _, user := range []struct{ name, role string }{{"admin", "admin"}, {"ops", "runner"}, {"guest", "viewer"}} {
		hash, err := access.HashPassword(user.name)
		if err != nil {
			t.Fatal(err)
		}
		users += "  - name: " + user.name + "\n    password: \"" + hash + "\"\n    role: " + user.role + "\n"
	}
	mkfile(t, filepath.Join(root, "task-ui.yml"), users+`projects:
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
	return root
}

func testConfig(t *testing.T, root string) *Config {
	t.Helper()
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
	return cfg
}

// testServer serves the test root.
func testServer(t *testing.T) (*httptest.Server, *Server) {
	t.Helper()
	svc, err := NewWithConfig(os.DirFS(".."), testConfig(t, testRoot(t)))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(svc.Handler())
	t.Cleanup(ts.Close)
	return ts, svc
}

func client() *http.Client {
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
	c := client()
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

// post submits a form and returns the status and the Location header.
func post(t *testing.T, c *http.Client, u string, form url.Values) (int, string) {
	t.Helper()
	resp, err := c.PostForm(u, form)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	return resp.StatusCode, resp.Header.Get("Location")
}

func TestNew(t *testing.T) {
	root := testRoot(t)
	t.Setenv("TASK_UI_ROOT", root)
	t.Setenv("TASK_UI_LISTEN", "127.0.0.1:0")
	svc, err := New(os.DirFS(".."), []string{"--history-output", filepath.Join(t.TempDir(), "h")})
	if err != nil {
		t.Fatal(err)
	}
	if svc.config.Root != root || !svc.hasUsers() {
		t.Fatalf("config: %+v", svc.config)
	}
	if _, err := New(os.DirFS(".."), []string{"--no-such-flag"}); err == nil {
		t.Fatal("unknown flag accepted")
	}
}

func TestServer_Start(t *testing.T) {
	svc, err := NewWithConfig(os.DirFS(".."), testConfig(t, testRoot(t)))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- svc.Start(ctx) }()
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after cancel")
	}

	svc.config.Listen = "nonsense"
	if err := svc.Start(context.Background()); err == nil {
		t.Fatal("invalid listen address accepted")
	}
}
