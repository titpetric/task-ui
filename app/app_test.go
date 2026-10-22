package app_test

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// newApp builds an instance over a root with one folder, an in-memory
// database and the users file, signed in as the given user.
func newApp(t *testing.T, autoLogin string) *app.App {
	t.Helper()
	root := t.TempDir()
	makeFolder(t, root, "site")
	cfg := &app.Config{
		Listen:     "127.0.0.1:0",
		Root:       root,
		Discover:   true,
		Htpasswd:   htpasswdFile(t, "admin", "guest"),
		Admins:     []string{"admin"},
		Viewers:    []string{"guest"},
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
	assert.NotNil(t, a.UsersFile)
	assert.Equal(t, "", a.Addr())

	// The users file seeds storage, and discovery finds the folder.
	admin, err := a.Users.GetByName(context.Background(), "admin")
	assert.NoError(t, err)
	assert.Equal(t, model.RoleAdmin, admin.Role)
	assert.Equal(t, "site", a.Catalogue.Find("site").ID)
}

func TestNew_Errors(t *testing.T) {
	t.Parallel()

	broken := filepath.Join(t.TempDir(), ".htpasswd")
	assert.NoError(t, os.WriteFile(broken, []byte("admin:{SHA}x\n"), 0o600))
	bad := &app.Config{Root: t.TempDir(), DataDir: t.TempDir(), DB: ":memory:", SessionTTL: time.Hour, Htpasswd: broken}
	_, err := app.New(context.Background(), bad, "test")
	assert.ErrorContains(t, err, "htpasswd -B")

	// A missing password file still starts the instance, which reads
	// the file once it appears.
	missing := filepath.Join(t.TempDir(), ".htpasswd")
	later := &app.Config{Root: t.TempDir(), DataDir: t.TempDir(), DB: ":memory:", SessionTTL: time.Hour, Htpasswd: missing}
	b, err := app.New(context.Background(), later, "test")
	assert.NoError(t, err)
	assert.True(t, !b.UsersFile.Changed())
	hash, _ := access.HashPassword("first")
	assert.NoError(t, access.SetHtpasswd(missing, "first", hash))
	assert.True(t, b.UsersFile.Changed())
	created, _, err := b.UsersFile.Sync(context.Background(), b.Users, b.Passwords)
	assert.NoError(t, err)
	assert.Equal(t, 1, created)

	// Without a file the instance starts with whoever is in storage, and
	// the instance makes the auto login user so development works from nothing.
	none := &app.Config{Root: t.TempDir(), DataDir: t.TempDir(), DB: ":memory:", SessionTTL: time.Hour, AutoLogin: "dev"}
	a, err := app.New(context.Background(), none, "test")
	assert.NoError(t, err)
	assert.Nil(t, a.UsersFile)
	dev, err := a.Users.GetByName(context.Background(), "dev")
	assert.NoError(t, err)
	assert.Equal(t, model.RoleAdmin, dev.Role)
	status, _, _ := get(t, a.Handler(), "/")
	assert.Equal(t, http.StatusOK, status)
	// The default workspace has its key directories from the start and
	// no encryption key: an admin adds one before it stores a secret.
	root, err := storage.NewWorkspaceStorage(a.DB).GetByFolder(context.Background(), model.Default)
	assert.NoError(t, err)
	assert.Nil(t, none.Keyring().EncryptKey(root.ID))
	assert.DirExists(t, none.Keyring().Encrypt(root.ID).Dir)
	a.Close()

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

	status, _, _ = get(t, signedIn.Handler(), "/f/_/site")
	assert.Equal(t, http.StatusOK, status)
	// The default workspace is there from the start, and holds site.
	status, body, _ = get(t, signedIn.Handler(), "/api/workspaces")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, body, `"folder":"_"`)
	status, body, _ = get(t, signedIn.Handler(), "/f/_")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, body, `href="/f/_/site"`)

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
	a.UsersFile.Interval = 20 * time.Millisecond

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

	// The instance watches the password file while serving: a new user signs in
	// without a restart.
	hash, _ := access.HashPassword("demo")
	assert.NoError(t, access.SetHtpasswd(a.Config.Htpasswd, "late", hash))
	assert.NoError(t, os.Chtimes(a.Config.Htpasswd, time.Now(), time.Now().Add(time.Second)))
	assert.Eventually(t, func() bool {
		_, err := a.Users.GetByName(context.Background(), "late")
		return err == nil
	}, 5*time.Second, 20*time.Millisecond)

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

// An MCP token made from the settings signs in to /api/mcp; a rotation
// ends the old one at once, and the CLI's token, without an expiry,
// keeps working throughout.
func TestApp_MCPToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a := newApp(t, "")
	handler := a.Handler()
	admin, err := a.Users.GetByName(ctx, "admin")
	assert.NoError(t, err)
	guest, err := a.Users.GetByName(ctx, "guest")
	assert.NoError(t, err)
	cli, hash, err := access.NewToken()
	assert.NoError(t, err)
	assert.NoError(t, a.Guard.Tokens.Set(ctx, admin.ID, "cli", hash))

	call := func(token, method, path, body string) (int, string) {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		raw, _ := io.ReadAll(rec.Body)
		return rec.Code, string(raw)
	}
	initialize := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}`
	enable := func() string {
		status, raw := call(cli, http.MethodPost, "/api/users/"+guest.ID+"/mcp", "")
		assert.Equal(t, http.StatusCreated, status, raw)
		var answer struct {
			Token     string `json:"token"`
			ExpiresAt string `json:"expires_at"`
		}
		assert.NoError(t, json.Unmarshal([]byte(raw), &answer))
		assert.True(t, strings.HasPrefix(answer.Token, access.TokenPrefix))
		assert.NotEqual(t, "", answer.ExpiresAt)
		return answer.Token
	}

	first := enable()
	status, _ := call(first, http.MethodPost, "/api/mcp", initialize)
	assert.Equal(t, http.StatusOK, status)

	second := enable()
	status, _ = call(first, http.MethodPost, "/api/mcp", initialize)
	assert.Equal(t, http.StatusUnauthorized, status)
	status, _ = call(second, http.MethodPost, "/api/mcp", initialize)
	assert.Equal(t, http.StatusOK, status)
	status, _ = call(cli, http.MethodPost, "/api/mcp", initialize)
	assert.Equal(t, http.StatusOK, status)

	// The listing tells MCP is on, with no token.
	status, raw := call(cli, http.MethodGet, "/api/users", "")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, raw, `"mcp":{"enabled":true`)
	assert.NotContains(t, raw, second)
	assert.NotContains(t, raw, access.HashToken(second))
	assert.NotContains(t, raw, access.TokenPrefix)

	status, _ = call(cli, http.MethodDelete, "/api/users/"+guest.ID+"/mcp", "")
	assert.Equal(t, http.StatusNoContent, status)
	status, _ = call(second, http.MethodPost, "/api/mcp", initialize)
	assert.Equal(t, http.StatusUnauthorized, status)
}
