package frontend_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestUsers(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	body := f.page(t, asAdmin, "/settings/users")
	assert.Contains(t, body, "<h2>Users</h2>")
	assert.Contains(t, body, "task-ui passwd")
	assert.Contains(t, body, "TASK_UI_ADMINS")
	for _, name := range []string{asAdmin, asOps, asViewer, asLimited} {
		assert.Contains(t, body, `href="/settings/users/`+f.users[name].ID+`"`)
	}
	// The page removes no user, and makes one only with a password file
	// to write.
	assert.NotContains(t, body, "data-delete")
	assert.NotContains(t, body, "/settings/users/new")
	assert.Contains(t, body, "infra</td>")
}

// The users page tells how to connect an MCP client, with the base link
// configured or derived from the proxy's forwarded headers, and never a
// token.
func TestUsers_MCP(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	req, err := http.NewRequest(http.MethodGet, f.server.URL+"/settings/users", nil)
	assert.NoError(t, err)
	session, err := f.opts.Guard.Sessions.Issue(context.Background(), f.users[asAdmin])
	assert.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: access.Cookie, Value: session})
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "task.example.com")
	res, err := client().Do(req)
	assert.NoError(t, err)
	raw, err := io.ReadAll(res.Body)
	assert.NoError(t, res.Body.Close())
	assert.NoError(t, err)
	derived := string(raw)
	assert.Contains(t, derived, "Connect an MCP client")
	assert.Contains(t, derived, `data-copy="https://task.example.com"`)
	assert.Contains(t, derived, `data-copy="https://task.example.com/api/mcp"`)
	assert.Contains(t, derived, `href="/settings/users/`+f.users[asAdmin].ID+`#mcp">your user settings</a>`)
	assert.NotContains(t, derived, access.TokenPrefix)

	f.opts.BaseURL = "https://task.cdn.si"
	h, err := frontend.New(f.opts)
	assert.NoError(t, err)
	f.server.Config.Handler = h.Routes()
	configured := f.page(t, asAdmin, "/settings/users")
	assert.Contains(t, configured, `data-copy="https://task.cdn.si"`)
	assert.Contains(t, configured, `data-copy="https://task.cdn.si/api/mcp"`)
	assert.Contains(t, configured, "claude mcp add --transport http --scope user task https://task.cdn.si/api/mcp")
	assert.NotContains(t, configured, "task.example.com")
	assert.NotContains(t, configured, access.TokenPrefix)
}

func TestUserForm(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	edit := f.page(t, asAdmin, "/settings/users/"+f.users[asLimited].ID)
	assert.Contains(t, edit, "<h2>Folders of limited</h2>")
	assert.Contains(t, edit, `data-api="/api/users/`+f.users[asLimited].ID+`" data-method="PUT"`)
	// The name and the role are text, not fields; the folders are Custom,
	// infra checked.
	assert.NotContains(t, edit, `id="u-name"`)
	assert.Contains(t, edit, `<span class="badge">`+model.RoleRunner+`</span>`)
	assert.Contains(t, edit, `<input type="radio" name="scope" value="custom" checked>`)
	assert.Contains(t, edit, `name="folders" value="infra" data-type="set" checked>`)
	assert.NotContains(t, edit, `name="password"`)
	assert.NotContains(t, edit, `name="role"`)
	// All projects: the boxes go off, none sent.
	all := f.page(t, asAdmin, "/settings/users/"+f.users[asOps].ID)
	assert.Contains(t, all, `<input type="radio" name="scope" value="all" checked>`)
	assert.Contains(t, all, `name="folders" value="infra" data-type="set" disabled>`)

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/users/"+model.NewID()).Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/settings/users/new").Status)

	// With a password file to write, Create user makes one.
	f.opts.AddUsers = true
	h, err := frontend.New(f.opts)
	assert.NoError(t, err)
	f.server.Config.Handler = h.Routes()
	assert.Contains(t, f.page(t, asAdmin, "/settings/users"), `href="/settings/users/new">Create user</a>`)
	created := f.page(t, asAdmin, "/settings/users/new")
	assert.Contains(t, created, `data-api="/api/users" data-method="POST"`)
	assert.Contains(t, created, `name="password" required minlength="8"`)
}

// A user's page shows their MCP access: Enable for user/edit, the dates
// with Rotate and Disable once on, for its own user too, and never the
// token. The expiry turns amber within a week and red when past.
func TestUserForm_MCP(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ops := f.users[asOps]
	own := "/settings/users/" + ops.ID

	off := f.page(t, asAdmin, own)
	assert.Contains(t, off, "MCP access")
	assert.Contains(t, off, `data-api="/api/users/`+ops.ID+`/mcp" data-method="POST" data-reveal="#mcp-token"`)
	assert.Contains(t, off, "Enable MCP</button>")
	assert.Contains(t, off, `data-fill="claude mcp add --transport http --scope user task http://127.0.0.1`)

	// Their own page opens for ops without user/view, with no folders
	// form and no Enable; another user's page stays closed.
	mine := f.page(t, asOps, own)
	assert.Contains(t, mine, "MCP access")
	assert.Contains(t, mine, "ask an admin")
	assert.NotContains(t, mine, "Enable MCP</button>")
	assert.NotContains(t, mine, `name="folders"`)
	assert.Equal(t, http.StatusForbidden, f.get(t, asOps, "/settings/users/"+f.users[asViewer].ID).Status)

	token, hash, err := access.NewToken()
	assert.NoError(t, err)
	expires := f.opts.Now().Add(access.MCPTokenTTL)
	_, err = f.opts.Guard.Tokens.SetUntil(context.Background(), ops.ID, access.MCPToken, hash, expires)
	assert.NoError(t, err)
	on := f.page(t, asOps, own)
	assert.Contains(t, on, "Rotate</button>")
	assert.Contains(t, on, `data-delete="/api/users/`+ops.ID+`/mcp"`)
	assert.Contains(t, on, `<span class="badge" title=`)
	assert.NotContains(t, on, token)
	assert.NotContains(t, on, hash)
	assert.NotContains(t, f.page(t, asAdmin, own), token)

	_, err = f.opts.Guard.Tokens.SetUntil(context.Background(), ops.ID, access.MCPToken, hash, f.opts.Now().Add(3*24*time.Hour))
	assert.NoError(t, err)
	assert.Contains(t, f.page(t, asOps, own), `<span class="badge signal" title=`)
	_, err = f.opts.Guard.Tokens.SetUntil(context.Background(), ops.ID, access.MCPToken, hash, f.opts.Now().Add(-time.Hour))
	assert.NoError(t, err)
	expired := f.page(t, asOps, own)
	assert.Contains(t, expired, `<span class="badge bad" title=`)
	assert.Contains(t, expired, ">expired ")
}
