package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// htpasswdFile writes a password file with one line per name, every
// password "demo", and returns its path.
func htpasswdFile(t *testing.T, names ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), ".htpasswd")
	hash, err := access.HashPassword("demo")
	assert.NoError(t, err)
	content := ""
	for _, name := range names {
		content += name + ":" + hash + "\n"
	}
	assert.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func newUsers(t *testing.T) *storage.UserStorage {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return storage.NewUserStorage(db)
}

func TestNewUsersFile(t *testing.T) {
	t.Parallel()

	f := app.NewUsersFile("/etc/task-ui/.htpasswd", []string{"admin"}, []string{"guest", "admin"})
	assert.Equal(t, "/etc/task-ui/.htpasswd", f.Path)
	assert.Equal(t, 5*time.Second, f.Interval)
	assert.Equal(t, model.RoleAdmin, f.Role("admin"))
	assert.Equal(t, model.RoleViewer, f.Role("guest"))
	assert.Equal(t, model.RoleRunner, f.Role("ops"))
	assert.False(t, f.Changed())
}

func TestUsersFile_Sync(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users := newUsers(t)
	path := htpasswdFile(t, "admin", "ops", "guest")
	f := app.NewUsersFile(path, []string{"admin"}, []string{"guest"})

	assert.True(t, f.Changed())
	created, removed, err := f.Sync(ctx, users)
	assert.NoError(t, err)
	assert.Equal(t, 3, created)
	assert.Equal(t, 0, removed)
	assert.False(t, f.Changed())

	admin := mustUser(t, users, "admin")
	assert.Equal(t, model.RoleAdmin, admin.Role)
	assert.True(t, access.CheckPassword(admin, "demo"))
	assert.Equal(t, model.RoleRunner, mustUser(t, users, "ops").Role)
	assert.Equal(t, model.RoleViewer, mustUser(t, users, "guest").Role)

	// A changed file: ops gets a new password and leaves the viewers,
	// guest is gone, a newcomer arrives; folders set in the UI stay.
	ops := mustUser(t, users, "ops")
	ops.Folders = model.EncodeList([]string{"infra"})
	assert.NoError(t, users.Update(ctx, ops))
	reset, _ := access.HashPassword("changed")
	assert.NoError(t, access.SetHtpasswd(path, "ops", reset))
	assert.NoError(t, access.SetHtpasswd(path, "dev", reset))
	data, _ := os.ReadFile(path)
	assert.NoError(t, os.WriteFile(path, []byte(replaceLine(string(data), "guest")), 0o600))
	f.Viewers = []string{"ops"}

	created, removed, err = f.Sync(ctx, users)
	assert.NoError(t, err)
	assert.Equal(t, 1, created)
	assert.Equal(t, 1, removed)
	ops = mustUser(t, users, "ops")
	assert.True(t, access.CheckPassword(ops, "changed"))
	assert.Equal(t, model.RoleViewer, ops.Role)
	assert.Equal(t, []string{"infra"}, ops.FolderList())
	assert.Equal(t, admin.ID, mustUser(t, users, "admin").ID)
	_, err = users.GetByName(ctx, "guest")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// A missing or broken file syncs nothing.
	f.Path = filepath.Join(t.TempDir(), "missing")
	_, _, err = f.Sync(ctx, users)
	assert.Error(t, err)
	assert.NoError(t, os.WriteFile(path, []byte("broken\n"), 0o600))
	f.Path = path
	_, _, err = f.Sync(ctx, users)
	assert.Error(t, err)
	assert.Len(t, mustList(t, users), 3)
}

func TestUsersFile_Watch(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	users := newUsers(t)
	path := htpasswdFile(t, "admin")
	f := app.NewUsersFile(path, []string{"admin"}, nil)
	f.Interval = 20 * time.Millisecond
	_, _, err := f.Sync(ctx, users)
	assert.NoError(t, err)

	errs := make(chan error, 4)
	done := make(chan struct{})
	go func() {
		f.Watch(ctx, users, func(err error) { errs <- err })
		close(done)
	}()

	// A new line shows up without a restart.
	hash, _ := access.HashPassword("demo")
	time.Sleep(10 * time.Millisecond)
	assert.NoError(t, access.SetHtpasswd(path, "ops", hash))
	assert.NoError(t, os.Chtimes(path, time.Now(), time.Now().Add(time.Second)))
	assert.Eventually(t, func() bool { return len(mustList(t, users)) == 2 }, 5*time.Second, 20*time.Millisecond)

	// A broken file is reported and changes nothing.
	assert.NoError(t, os.WriteFile(path, []byte("broken\n"), 0o600))
	assert.NoError(t, os.Chtimes(path, time.Now(), time.Now().Add(2*time.Second)))
	select {
	case err := <-errs:
		assert.ErrorContains(t, err, "line 1")
	case <-time.After(5 * time.Second):
		t.Fatal("the broken file was not reported")
	}
	assert.Len(t, mustList(t, users), 2)

	cancel()
	<-done
}

func mustUser(t *testing.T, users *storage.UserStorage, name string) *model.User {
	t.Helper()
	user, err := users.GetByName(context.Background(), name)
	assert.NoError(t, err)
	return user
}

func mustList(t *testing.T, users *storage.UserStorage) []model.User {
	t.Helper()
	list, err := users.List(context.Background())
	assert.NoError(t, err)
	return list
}

// replaceLine drops the line of a user from a password file.
func replaceLine(content, name string) string {
	var out string
	for _, line := range splitLines(content) {
		if line != "" && !hasPrefix(line, name+":") {
			out += line + "\n"
		}
	}
	return out
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
