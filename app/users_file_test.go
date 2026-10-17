package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// usersYAML writes a users file and returns its path.
func usersYAML(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "task-ui.yml")
	assert.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestLoadUsersFile(t *testing.T) {
	t.Parallel()

	file, err := app.LoadUsersFile(usersYAML(t, "users:\n  - name: admin\n    password: demo\n    role: admin\n  - name: guest\n    password: demo\n    folders: [infra]\n"))
	assert.NoError(t, err)
	assert.Len(t, file.Users, 2)
	assert.Equal(t, model.RoleAdmin, file.Users[0].Role)
	assert.Equal(t, []string{"infra"}, file.Users[1].Folders)

	_, err = app.LoadUsersFile(filepath.Join(t.TempDir(), "missing.yml"))
	assert.Error(t, err)
	_, err = app.LoadUsersFile(usersYAML(t, "users: [\n"))
	assert.Error(t, err)
	_, err = app.LoadUsersFile(usersYAML(t, "users:\n  - name: admin\n    password: x\n  - name: admin\n    password: y\n"))
	assert.ErrorContains(t, err, "listed twice")
	_, err = app.LoadUsersFile(usersYAML(t, "users:\n  - name: admin\n    password: x\n    role: king\n"))
	assert.ErrorContains(t, err, "king")
}

func TestUsersFile_Seed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users := storage.NewUserStorage(db)

	file := &app.UsersFile{Users: []app.UserEntry{{Name: "admin", Password: "demo", Role: model.RoleAdmin}, {Name: "ops", Password: "demo", Role: model.RoleRunner}}}
	assert.NoError(t, file.Validate())
	created, err := file.Seed(ctx, users)
	assert.NoError(t, err)
	assert.Equal(t, 2, created)

	admin, err := users.GetByName(ctx, "admin")
	assert.NoError(t, err)
	assert.True(t, access.CheckPassword(admin, "demo"))

	// Seeding again changes what the file says and creates nothing.
	file.Users[1].Role, file.Users[1].Password = model.RoleViewer, "changed"
	created, err = file.Seed(ctx, users)
	assert.NoError(t, err)
	assert.Equal(t, 0, created)
	ops, err := users.GetByName(ctx, "ops")
	assert.NoError(t, err)
	assert.Equal(t, model.RoleViewer, ops.Role)
	assert.True(t, access.CheckPassword(ops, "changed"))
	assert.Equal(t, admin.ID, mustUser(t, users, "admin").ID)
}

func mustUser(t *testing.T, users *storage.UserStorage, name string) *model.User {
	t.Helper()
	user, err := users.GetByName(context.Background(), name)
	assert.NoError(t, err)
	return user
}
