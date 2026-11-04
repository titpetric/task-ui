//go:build integration

package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
	"github.com/titpetric/task-ui/storage"
)

func TestUserStorage_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))

	user := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, user))
	assert.True(t, model.ValidID(user.ID))
	assert.NotNil(t, user.CreatedAt)
	assert.Equal(t, "[]", user.Folders)

	assert.Error(t, users.Create(ctx, &model.User{Role: model.RoleAdmin}), "no name")
	assert.Error(t, users.Create(ctx, &model.User{Name: "x", Role: "root"}), "bad role")
	assert.Error(t, users.Create(ctx, &model.User{Name: "ops", Role: model.RoleAdmin}), "duplicate name")
}

func TestUserStorage_Get(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))
	user := &model.User{Name: "admin", Role: model.RoleAdmin, Folders: `["web"]`}
	assert.NoError(t, users.Create(ctx, user))

	got, err := users.Get(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, "admin", got.Name)
	assert.Equal(t, []string{"web"}, got.FolderList())
	assert.NotNil(t, got.CreatedAt)

	_, err = users.Get(ctx, "missing")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestUserStorage_GetByName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))
	assert.NoError(t, users.Create(ctx, &model.User{Name: "guest", Role: model.RoleViewer}))

	got, err := users.GetByName(ctx, "guest")
	assert.NoError(t, err)
	assert.Equal(t, model.RoleViewer, got.Role)
	_, err = users.GetByName(ctx, "nobody")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestUserStorage_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))
	for _, name := range []string{"zed", "amy"} {
		assert.NoError(t, users.Create(ctx, &model.User{Name: name, Role: model.RoleViewer}))
	}
	list, err := users.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
	assert.Equal(t, "amy", list[0].Name)

	count, err := users.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestUserStorage_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))
	user := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, user))

	user.Role = model.RoleAdmin
	user.Folders = model.EncodeList([]string{"infra"})
	assert.NoError(t, users.Update(ctx, user))
	got, err := users.Get(ctx, user.ID)
	assert.NoError(t, err)
	assert.True(t, got.Admin, "the admin role is the admin flag")
	assert.Equal(t, model.RoleRunner, got.Role)
	assert.Equal(t, model.RoleAdmin, got.EffectiveRole())
	assert.Equal(t, []string{"infra"}, got.FolderList())

	got.AgentReply = true
	assert.ErrorIs(t, users.Update(ctx, got), storage.ErrAgentReply)

	user.Role = "root"
	assert.Error(t, users.Update(ctx, user))
}

// Deleting a user ends their sessions with them.
func TestUserStorage_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	users, sessions := storage.NewUserStorage(db), storage.NewUserSessionStorage(db)
	user := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, user))
	_, err := sessions.Create(ctx, user.ID, "h1", time.Now().Add(time.Hour))
	assert.NoError(t, err)
	tokens := storage.NewUserAuthTokenStorage(db)
	assert.NoError(t, tokens.Set(ctx, user.ID, "cli", "t1"))
	assert.NoError(t, users.Policies().Put(ctx, permission.KindUser, "ops", "note/view", permission.Deny))
	assert.False(t, users.Permissions(ctx, user).Allowed("note/view"))
	assert.NoError(t, users.Delete(ctx, user.ID))
	_, err = users.Get(ctx, user.ID)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = sessions.Find(ctx, "h1")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	_, err = tokens.Find(ctx, "t1")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	policy, err := users.Policies().Policy(ctx, permission.KindUser, "ops")
	assert.NoError(t, err)
	assert.Len(t, policy, 0)
	assert.NoError(t, users.Delete(ctx, user.ID), "deleting twice is fine")
}

func TestUserStorage_MoveFolder(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := storage.NewUserStorage(newDB(t))
	limited := &model.User{Name: "dev", Role: model.RoleRunner, Folders: model.EncodeList([]string{"web", "web/docs", "website"})}
	everyone := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, limited))
	assert.NoError(t, users.Create(ctx, everyone))

	changed, err := users.MoveFolder(ctx, "web", "acme/web")
	assert.NoError(t, err)
	assert.Equal(t, 1, changed)
	got, err := users.Get(ctx, limited.ID)
	assert.NoError(t, err)
	assert.Equal(t, []string{"acme/web", "acme/web/docs", "website"}, got.FolderList())
	got, err = users.Get(ctx, everyone.ID)
	assert.NoError(t, err)
	assert.Len(t, got.FolderList(), 0, "a user who sees every folder keeps seeing them")
}

// SetFlags changes what a user is; the last admin keeps the flag unless
// forced, MCP off ends agent reply's use and the user's MCP token.
func TestUserStorage_SetFlags(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	users, tokens := storage.NewUserStorage(db), storage.NewUserAuthTokenStorage(db)
	root := &model.User{Name: "root", Role: model.RoleAdmin}
	ops := &model.User{Name: "ops", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, root))
	assert.NoError(t, users.Create(ctx, ops))
	assert.True(t, root.Admin)
	n, err := users.Admins(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)

	assert.ErrorIs(t, users.SetFlags(ctx, root, false, model.RoleRunner, false, false, false), storage.ErrLastAdmin)
	assert.Error(t, users.SetFlags(ctx, ops, false, model.RoleAdmin, false, false, false), "admin is the flag")
	assert.ErrorIs(t, users.SetFlags(ctx, ops, false, model.RoleRunner, false, true, false), storage.ErrAgentReply)

	assert.NoError(t, users.SetFlags(ctx, ops, true, model.RoleViewer, true, true, false))
	assert.True(t, ops.Admin && ops.Mcp && ops.AgentReply)
	agents, err := users.Agents(ctx)
	assert.NoError(t, err)
	assert.Len(t, agents, 1)
	assert.NoError(t, users.SetFlags(ctx, root, false, model.RoleRunner, false, false, false), "ops is an admin too")

	expires := time.Now().Add(time.Hour)
	_, err = tokens.Add(ctx, ops.ID, model.TokenMCP, "mcp", "h1", &expires)
	assert.NoError(t, err)
	assert.NoError(t, tokens.Set(ctx, ops.ID, "cli", "h2"))
	assert.NoError(t, users.SetFlags(ctx, ops, true, model.RoleViewer, false, false, false))
	list, err := tokens.List(ctx, ops.ID)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "cli", list[0].Name, "the CLI token stays; the endpoint refuses it")

	assert.NoError(t, users.SetFlags(ctx, ops, false, model.RoleViewer, false, false, true), "forced")
	n, _ = users.Admins(ctx)
	assert.Equal(t, int64(0), n)
}

// Every method hands back what the driver said.
func TestUserStorage_Closed(t *testing.T) {
	t.Parallel()
	users := storage.NewUserStorage(closedDB(t))
	ctx := context.Background()
	user := &model.User{Name: "alice", Role: model.RoleRunner}

	assert.ErrorContains(t, users.Create(ctx, user), errClosed)
	_, err := users.Get(ctx, "u1")
	assert.ErrorContains(t, err, errClosed)
	_, err = users.GetByName(ctx, "alice")
	assert.ErrorContains(t, err, errClosed)
	_, err = users.List(ctx)
	assert.ErrorContains(t, err, errClosed)
	_, err = users.Count(ctx)
	assert.ErrorContains(t, err, errClosed)
	_, err = users.Admins(ctx)
	assert.ErrorContains(t, err, errClosed)
	_, err = users.Agents(ctx)
	assert.ErrorContains(t, err, errClosed)
	assert.ErrorContains(t, users.Update(ctx, user), errClosed)
	assert.ErrorContains(t, users.SetFlags(ctx, user, true, model.RoleRunner, true, false, false), errClosed)
	assert.ErrorContains(t, users.Delete(ctx, "u1"), errClosed)
	_, err = users.MoveFolder(ctx, "acme", "acme2")
	assert.ErrorContains(t, err, errClosed)
}

// SetFlags counts the admins before it takes the flag from the last of
// them, writes the user back, and drops the MCP token when MCP goes: it
// stops at whichever of those failed.
func TestUserStorage_SetFlagsErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	counting := newDB(t)
	users := storage.NewUserStorage(counting)
	admin := &model.User{Name: "root", Role: model.RoleAdmin}
	assert.NoError(t, users.Create(ctx, admin))
	second := &model.User{Name: "other", Role: model.RoleAdmin}
	assert.NoError(t, users.Create(ctx, second))

	refuse(t, counting, model.UserTable, "UPDATE")
	assert.ErrorContains(t, users.SetFlags(ctx, second, false, model.RoleRunner, true, false, false), errRefused)

	tokens := newDB(t)
	withToken := storage.NewUserStorage(tokens)
	user := &model.User{Name: "alice", Role: model.RoleRunner, Mcp: true}
	assert.NoError(t, withToken.Create(ctx, user))
	expires := time.Now().Add(time.Hour)
	_, err := storage.NewUserAuthTokenStorage(tokens).Add(ctx, user.ID, model.TokenMCP, "mcp", "hash", &expires)
	assert.NoError(t, err)
	refuse(t, tokens, model.UserAuthTokenTable, "DELETE")
	assert.ErrorContains(t, withToken.SetFlags(ctx, user, false, model.RoleRunner, false, false, false), errRefused)
}

// Delete clears what a user signs in with, its policy and then its row,
// and stops at whichever of those failed.
func TestUserStorage_DeleteErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	for _, table := range []string{model.UserEmailTable, model.UserSessionTable, model.UserAuthPasswordTable, model.UserAuthTokenTable, model.UserTable} {
		db := brokenDB(t, table)
		assert.ErrorContains(t, storage.NewUserStorage(db).Delete(ctx, "u1"), errNoTable, table)
	}

	policies := newDB(t)
	users := storage.NewUserStorage(policies)
	user := &model.User{Name: "alice", Role: model.RoleRunner}
	assert.NoError(t, users.Create(ctx, user))
	assert.NoError(t, users.Policies().Put(ctx, permission.KindUser, "alice", "run", permission.Allow))
	refuse(t, policies, model.PermissionPolicyTable, "DELETE")
	assert.ErrorContains(t, users.Delete(ctx, user.ID), errRefused)
}

// A move rewrites the folder list of a user at a time, and stops at the
// user that failed.
func TestUserStorage_MoveFolderErrors(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db := newDB(t)
	users := storage.NewUserStorage(db)
	assert.NoError(t, users.Create(ctx, &model.User{Name: "alice", Role: model.RoleRunner, Folders: model.EncodeList([]string{"acme"})}))

	refuse(t, db, model.UserTable, "UPDATE")
	_, err := users.MoveFolder(ctx, "acme", "acme2")
	assert.ErrorContains(t, err, errRefused)
}
