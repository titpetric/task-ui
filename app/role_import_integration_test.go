//go:build integration

package app_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The role lists of an older configuration go into storage once, on
// the first start: admins hold the flag, viewers the role, the agent
// user MCP and agent reply. A later start reads them no more.
func TestNew_ImportsRolesOnce(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	makeFolder(t, root, "site")
	cfg := &app.Config{
		Listen: "127.0.0.1:0", Root: root, Htpasswd: htpasswdFile(t, "black", "claude", "guest", "ops"),
		Import:  &app.RoleImport{Admins: []string{"black"}, Viewers: []string{"guest"}, Agent: "claude"},
		DataDir: filepath.Join(t.TempDir(), "data"), SSHDir: t.TempDir(), SessionTTL: time.Hour, Keep: time.Minute, Grace: time.Second,
	}
	assert.NoError(t, cfg.Validate())
	ctx := context.Background()
	start := func() map[string]model.User {
		a, err := app.New(ctx, cfg, "test")
		assert.NoError(t, err)
		defer a.Close()
		list, err := a.Users.List(ctx)
		assert.NoError(t, err)
		users := map[string]model.User{}
		for _, u := range list {
			users[u.Name] = u
		}
		return users
	}
	users := start()
	assert.True(t, users["black"].Admin)
	assert.False(t, users["ops"].Admin, "the first read made everyone an admin; the list decides")
	assert.Equal(t, model.RoleRunner, users["ops"].Role)
	assert.Equal(t, model.RoleViewer, users["guest"].Role)
	assert.True(t, users["claude"].Mcp && users["claude"].AgentReply)
	assert.False(t, users["claude"].Admin)

	cfg.Import = &app.RoleImport{Admins: []string{"ops"}}
	users = start()
	assert.True(t, users["black"].Admin, "read once")
	assert.False(t, users["ops"].Admin)
}
