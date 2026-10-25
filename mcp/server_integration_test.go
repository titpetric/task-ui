//go:build integration

package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/storage"
)

// fixture is an instance with a folder, an admin and a viewer, each with
// a token, served over HTTP.
type fixture struct {
	server *httptest.Server
	tokens map[string]string
	root   string
	notes  *storage.NoteStorage
	board  *note.Board
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "site"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "site", "Makefile"), []byte("build: ## Build the site\n\techo built\ndefault: ## Check the site\n\techo checked\n"), 0o644))

	hash, err := access.HashPassword("demo")
	assert.NoError(t, err)
	htpasswd := filepath.Join(t.TempDir(), ".htpasswd")
	assert.NoError(t, os.WriteFile(htpasswd, []byte("admin:"+hash+"\nguest:"+hash+"\nops:"+hash+"\nconsuela:"+hash+"\nplain:"+hash+"\n"), 0o600))

	cfg := &app.Config{
		Listen: "127.0.0.1:0", Root: root, Discover: true, Htpasswd: htpasswd,
		Import:  &app.RoleImport{Admins: []string{"admin"}, Viewers: []string{"guest"}},
		DataDir: filepath.Join(t.TempDir(), "data"), DB: ":memory:", SSHDir: t.TempDir(),
		SessionTTL: time.Hour, Keep: time.Minute, Grace: time.Second,
	}
	assert.NoError(t, cfg.Validate())
	a, err := app.New(ctx, cfg, "test")
	assert.NoError(t, err)
	t.Cleanup(func() { a.Close() })

	f := &fixture{server: httptest.NewServer(a.Handler()), tokens: map[string]string{}, root: root, notes: a.Board.Notes, board: a.Board}
	t.Cleanup(f.server.Close)
	tokens := storage.NewUserAuthTokenStorage(a.DB)
	for _, name := range []string{"admin", "guest", "ops", "consuela", "plain"} {
		user, err := a.Users.GetByName(ctx, name)
		assert.NoError(t, err)
		token, hash, err := access.NewToken()
		assert.NoError(t, err)
		assert.NoError(t, tokens.Set(ctx, user.ID, "test", hash))
		f.tokens[name] = token
		if name != "plain" {
			user.Mcp = true
			assert.NoError(t, a.Users.Update(ctx, user))
		}
	}
	return f
}

// post sends a body to a path with a token and returns the status and body.
func (f *fixture) post(t *testing.T, token, path string, body any) (int, []byte) {
	t.Helper()
	return f.send(t, http.MethodPost, token, path, body)
}

// put sends a body to a path with PUT, as post does with POST.
func (f *fixture) put(t *testing.T, token, path string, body any) (int, []byte) {
	t.Helper()
	return f.send(t, http.MethodPut, token, path, body)
}

// send sends a body to a path with a token and returns the status and body.
func (f *fixture) send(t *testing.T, method, token, path string, body any) (int, []byte) {
	t.Helper()
	data, err := json.Marshal(body)
	assert.NoError(t, err)
	req, err := http.NewRequest(method, f.server.URL+path, bytes.NewReader(data))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	return res.StatusCode, raw
}

// reply is a JSON-RPC response as the tests read it.
type reply struct {
	Result struct {
		ProtocolVersion string `json:"protocolVersion"`
		Instructions    string `json:"instructions"`
		Resources       []struct {
			URI string `json:"uri"`
		} `json:"resources"`
		Contents []struct {
			MimeType string `json:"mimeType"`
			Text     string `json:"text"`
		} `json:"contents"`
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	} `json:"result"`
	Error *struct {
		Code int `json:"code"`
	} `json:"error"`
}

// rpc sends a JSON-RPC request as a user and decodes the reply.
func (f *fixture) rpc(t *testing.T, user, method string, params any) reply {
	t.Helper()
	status, raw := f.post(t, f.tokens[user], "/api/mcp", map[string]any{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
	assert.Equal(t, http.StatusOK, status, string(raw))
	var r reply
	assert.NoError(t, json.Unmarshal(raw, &r), string(raw))
	return r
}

// tool calls a tool as a user and returns its text and whether it failed.
func (f *fixture) tool(t *testing.T, user, name string, args any) (string, bool) {
	t.Helper()
	r := f.rpc(t, user, "tools/call", map[string]any{"name": name, "arguments": args})
	assert.Nil(t, r.Error)
	assert.Len(t, r.Result.Content, 1)
	return r.Result.Content[0].Text, r.Result.IsError
}

// Only a token of a user gets in.
func TestServer_Auth(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	body := map[string]any{"jsonrpc": "2.0", "id": 1, "method": "ping"}

	status, _ := f.post(t, "", "/api/mcp", body)
	assert.Equal(t, http.StatusUnauthorized, status)
	status, _ = f.post(t, "tui_wrong", "/api/mcp", body)
	assert.Equal(t, http.StatusUnauthorized, status)
	status, _ = f.post(t, f.tokens["admin"], "/api/mcp", body)
	assert.Equal(t, http.StatusOK, status)
	status, raw := f.post(t, f.tokens["plain"], "/api/mcp", body)
	assert.Equal(t, http.StatusForbidden, status, "no MCP agent user")
	assert.Contains(t, string(raw), "MCP is off")

	// The rest of the API takes the token too.
	req, _ := http.NewRequest(http.MethodGet, f.server.URL+"/api/folders", nil)
	req.Header.Set("Authorization", "Bearer "+f.tokens["guest"])
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

// The server speaks JSON-RPC over POST and answers what it cannot read.
func TestServer_ServeHTTP(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	r := f.rpc(t, "admin", "initialize", map[string]any{"protocolVersion": "2025-03-26"})
	assert.Equal(t, "2025-03-26", r.Result.ProtocolVersion)
	r = f.rpc(t, "admin", "initialize", map[string]any{"protocolVersion": "1999-01-01"})
	assert.Equal(t, "2025-06-18", r.Result.ProtocolVersion)
	// The instructions carry context.md, then the tools and whoami.
	assert.True(t, strings.HasPrefix(r.Result.Instructions, "# task-ui\n"))
	assert.Contains(t, r.Result.Instructions, "## Artifacts (`artifact`)")
	assert.Contains(t, r.Result.Instructions, "Call whoami to see what you may do")

	status, raw := f.post(t, f.tokens["admin"], "/api/mcp", map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
	assert.Equal(t, http.StatusAccepted, status)
	assert.Len(t, raw, 0)

	assert.Len(t, f.rpc(t, "admin", "tools/list", nil).Result.Tools, 18)
	assert.Equal(t, -32601, f.rpc(t, "admin", "prompts/list", nil).Error.Code)
	assert.Equal(t, -32602, f.rpc(t, "admin", "tools/call", map[string]any{"name": "nothing"}).Error.Code)

	status, raw = f.post(t, f.tokens["admin"], "/api/mcp", "not an object")
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(raw), `"code":-32700`)
	status, raw = f.post(t, f.tokens["admin"], "/api/mcp", map[string]any{"id": 2, "method": "ping"})
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(raw), `"code":-32600`)

	req, _ := http.NewRequest(http.MethodGet, f.server.URL+"/api/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+f.tokens["admin"])
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	res.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
}
