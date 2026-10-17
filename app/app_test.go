package app_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// newApp builds an instance over a root with one folder, an in-memory
// database and the users file, signed in as the given user.
func newApp(t *testing.T, autoLogin string) *app.App {
	t.Helper()
	root := t.TempDir()
	makeFolder(t, root, "site")
	users := usersYAML(t, "users:\n  - name: admin\n    password: demo\n    role: admin\n  - name: guest\n    password: demo\n    role: viewer\n")
	cfg := &app.Config{
		Listen:     "127.0.0.1:0",
		Root:       root,
		Discover:   true,
		UsersFile:  users,
		DataDir:    filepath.Join(t.TempDir(), "data"),
		DB:         ":memory:",
		SSHDir:     t.TempDir(),
		SessionTTL: time.Hour,
		AutoLogin:  autoLogin,
		Keep:       time.Minute,
		Grace:      time.Second,
	}
	assert.NoError(t, cfg.Validate())
	a, err := app.New(context.Background(), cfg, "test")
	assert.NoError(t, err)
	t.Cleanup(func() { a.Close() })
	return a
}

func get(t *testing.T, handler http.Handler, path string) (int, string, http.Header) {
	t.Helper()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	body, _ := io.ReadAll(rec.Body)
	return rec.Code, string(body), rec.Header()
}

func TestNew(t *testing.T) {
	t.Parallel()
	a := newApp(t, "admin")

	assert.NotNil(t, a.DB)
	assert.NotNil(t, a.Runner)
	assert.NotNil(t, a.Scheduler)
	assert.NotNil(t, a.Guard)
	assert.NotNil(t, a.Handler())
	assert.Equal(t, "", a.Addr())

	// The users file was seeded, the folder discovered.
	admin, err := a.Users.GetByName(context.Background(), "admin")
	assert.NoError(t, err)
	assert.Equal(t, model.RoleAdmin, admin.Role)
	assert.Equal(t, "site", a.Catalogue.Find("site").ID)
}

func TestNew_Errors(t *testing.T) {
	t.Parallel()

	bad := &app.Config{Root: t.TempDir(), DataDir: t.TempDir(), DB: ":memory:", SessionTTL: time.Hour, UsersFile: usersYAML(t, "users: [\n")}
	_, err := app.New(context.Background(), bad, "test")
	assert.Error(t, err)

	unwritable := &app.Config{Root: t.TempDir(), DataDir: "/proc/task-ui-nope", DB: ":memory:", SessionTTL: time.Hour}
	_, err = app.New(context.Background(), unwritable, "test")
	assert.Error(t, err)
}

func TestApp_Handler(t *testing.T) {
	t.Parallel()

	signedIn := newApp(t, "admin")
	status, body, _ := get(t, signedIn.Handler(), "/")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, body, "Dashboard")
	assert.Contains(t, body, "site")

	status, body, headers := get(t, signedIn.Handler(), "/api/runs")
	assert.Equal(t, http.StatusOK, status)
	assert.True(t, strings.HasPrefix(headers.Get("Content-Type"), "application/json"))
	assert.Contains(t, body, `"runs":[]`)

	status, _, _ = get(t, signedIn.Handler(), "/f/site")
	assert.Equal(t, http.StatusOK, status)

	// Without a session the pages redirect and the API refuses.
	anonymous := newApp(t, "")
	status, _, headers = get(t, anonymous.Handler(), "/")
	assert.Equal(t, http.StatusSeeOther, status)
	assert.Equal(t, "/login?back=%2F", headers.Get("Location"))
	status, _, _ = get(t, anonymous.Handler(), "/api/runs")
	assert.Equal(t, http.StatusUnauthorized, status)
	status, _, _ = get(t, anonymous.Handler(), "/login")
	assert.Equal(t, http.StatusOK, status)
	status, _, _ = get(t, anonymous.Handler(), "/static/task-ui.css")
	assert.Equal(t, http.StatusOK, status)
}

func TestApp_Serve(t *testing.T) {
	t.Parallel()
	a := newApp(t, "guest")

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Serve(ctx, ln) }()

	var res *http.Response
	assert.Eventually(t, func() bool {
		var err error
		res, err = http.Get("http://" + ln.Addr().String() + "/healthz")
		return err == nil
	}, 5*time.Second, 20*time.Millisecond)
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, ln.Addr().String(), a.Addr())

	cancel()
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("serve did not return after the context ended")
	}
}

func TestApp_Run(t *testing.T) {
	t.Parallel()
	a := newApp(t, "guest")

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.Run(ctx) }()
	assert.Eventually(t, func() bool { return a.Addr() != "" }, 5*time.Second, 20*time.Millisecond)
	cancel()
	assert.NoError(t, <-done)

	taken := newApp(t, "guest")
	taken.Config.Listen = "256.256.256.256:1"
	assert.Error(t, taken.Run(context.Background()))
}
