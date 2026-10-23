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

func newUsers(t *testing.T) (*storage.UserStorage, *storage.UserAuthPasswordStorage) {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return storage.NewUserStorage(db), storage.NewUserAuthPasswordStorage(db)
}

// signsIn reports whether a user's stored password is the one given.
func signsIn(t *testing.T, passwords *storage.UserAuthPasswordStorage, user *model.User, password string) bool {
	t.Helper()
	hash, err := passwords.PasswordHash(context.Background(), user.ID)
	assert.NoError(t, err)
	return access.CheckHash(hash, password)
}

func TestNewUsersFile(t *testing.T) {
	t.Parallel()

	f := app.NewUsersFile("/etc/task-ui/.htpasswd")
	assert.Equal(t, "/etc/task-ui/.htpasswd", f.Path)
	assert.Equal(t, 5*time.Second, f.Interval)
	assert.False(t, f.Changed())
}

func TestUsersFile_Sync(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := storage.OpenAndMigrate(ctx, ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	users, passwords := storage.NewUserStorage(db), storage.NewUserAuthPasswordStorage(db)
	sessions := access.NewSessions(storage.NewUserSessionStorage(db), users, time.Hour)
	path := htpasswdFile(t, "admin", "ops", "guest")
	f := app.NewUsersFile(path)

	assert.True(t, f.Changed())
	created, removed, err := f.Sync(ctx, users, passwords)
	assert.NoError(t, err)
	assert.Equal(t, 3, created)
	assert.Equal(t, 0, removed)
	assert.False(t, f.Changed())

	// The first read into empty storage makes every user an admin.
	admin := mustUser(t, users, "admin")
	assert.True(t, admin.Admin)
	assert.True(t, signsIn(t, passwords, admin, "demo"))
	assert.True(t, mustUser(t, users, "ops").Admin)
	assert.True(t, mustUser(t, users, "guest").Admin)
	for _, name := range []string{"ops", "guest"} {
		u := mustUser(t, users, name)
		assert.NoError(t, users.SetFlags(ctx, u, false, model.RoleRunner, false, false, false))
	}

	// Everyone signs in once.
	tokens := map[string]string{}
	for _, name := range []string{"admin", "ops", "guest"} {
		tokens[name], err = sessions.Issue(ctx, mustUser(t, users, name))
		assert.NoError(t, err)
	}

	// A changed file: ops gets a new password, guest leaves, a newcomer
	// arrives as a viewer; folders and roles set in the UI stay.
	ops := mustUser(t, users, "ops")
	ops.Folders = model.EncodeList([]string{"infra"})
	assert.NoError(t, users.Update(ctx, ops))
	reset, _ := access.HashPassword("changed")
	assert.NoError(t, access.SetHtpasswd(path, "ops", reset))
	assert.NoError(t, access.SetHtpasswd(path, "dev", reset))
	data, _ := os.ReadFile(path)
	assert.NoError(t, os.WriteFile(path, []byte(replaceLine(string(data), "guest")), 0o600))

	created, removed, err = f.Sync(ctx, users, passwords)
	assert.NoError(t, err)
	assert.Equal(t, 1, created)
	assert.Equal(t, 1, removed)
	ops = mustUser(t, users, "ops")
	assert.True(t, signsIn(t, passwords, ops, "changed"))
	assert.Equal(t, model.RoleRunner, ops.Role)
	dev := mustUser(t, users, "dev")
	assert.Equal(t, model.RoleViewer, dev.Role, "the lowest role")
	assert.False(t, dev.Admin || dev.Mcp || dev.AgentReply)
	assert.Equal(t, []string{"infra"}, ops.FolderList())
	assert.Equal(t, admin.ID, mustUser(t, users, "admin").ID)
	_, err = users.GetByName(ctx, "guest")
	assert.ErrorIs(t, err, storage.ErrNotFound)

	// The new password and the removal end those sessions; admin's stays.
	_, err = sessions.Resolve(ctx, tokens["ops"])
	assert.ErrorIs(t, err, access.ErrInvalidToken)
	_, err = sessions.Resolve(ctx, tokens["guest"])
	assert.ErrorIs(t, err, access.ErrInvalidToken)
	_, err = sessions.Resolve(ctx, tokens["admin"])
	assert.NoError(t, err)

	// A missing or broken file syncs nothing.
	f.Path = filepath.Join(t.TempDir(), "missing")
	_, _, err = f.Sync(ctx, users, passwords)
	assert.Error(t, err)
	assert.NoError(t, os.WriteFile(path, []byte("broken\n"), 0o600))
	f.Path = path
	_, _, err = f.Sync(ctx, users, passwords)
	assert.Error(t, err)
	assert.Len(t, mustList(t, users), 3)
}

func TestUsersFile_Watch(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	users, passwords := newUsers(t)
	path := htpasswdFile(t, "admin")
	f := app.NewUsersFile(path)
	f.Interval = 20 * time.Millisecond
	_, _, err := f.Sync(ctx, users, passwords)
	assert.NoError(t, err)

	errs := make(chan error, 4)
	done := make(chan struct{})
	go func() {
		f.Watch(ctx, users, passwords, func(err error) { errs <- err })
		close(done)
	}()

	// A new line shows up without a restart.
	hash, _ := access.HashPassword("demo")
	time.Sleep(10 * time.Millisecond)
	assert.NoError(t, access.SetHtpasswd(path, "ops", hash))
	assert.NoError(t, os.Chtimes(path, time.Now(), time.Now().Add(time.Second)))
	assert.Eventually(t, func() bool { return len(mustList(t, users)) == 2 }, 5*time.Second, 20*time.Millisecond)

	// Watch reports a broken file, and the file changes nothing.
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

// Add writes a new user to the file and storage at once; a name the file
// holds, a bad name and a short password each get an error.
func TestUsersFile_Add(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users, passwords := newUsers(t)
	f := app.NewUsersFile(htpasswdFile(t, "admin"))
	_, _, err := f.Sync(ctx, users, passwords)
	assert.NoError(t, err)

	assert.NoError(t, f.Add(ctx, users, passwords, "alice", "correct horse"))
	alice := mustUser(t, users, "alice")
	assert.Equal(t, model.RoleViewer, alice.Role, "a new user of a storage with users")
	assert.False(t, alice.Admin)
	assert.True(t, signsIn(t, passwords, alice, "correct horse"))
	assert.False(t, f.Changed())
	creds, err := access.LoadHtpasswd(f.Path)
	assert.NoError(t, err)
	assert.Len(t, creds, 2)

	assert.ErrorIs(t, f.Add(ctx, users, passwords, "alice", "another one"), app.ErrUserExists)
	assert.Error(t, f.Add(ctx, users, passwords, "Bad:Name", "correct horse"))
	assert.Error(t, f.Add(ctx, users, passwords, "bob", "short"))

	fresh := app.NewUsersFile(filepath.Join(t.TempDir(), ".htpasswd"))
	assert.NoError(t, fresh.Add(ctx, users, passwords, "carol", "correct horse"))
}

// Remove takes a user out of the file and storage at once; a name the
// file does not hold is not found.
func TestUsersFile_Remove(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users, passwords := newUsers(t)
	f := app.NewUsersFile(htpasswdFile(t, "admin", "ops"))
	_, _, err := f.Sync(ctx, users, passwords)
	assert.NoError(t, err)

	assert.NoError(t, f.Remove(ctx, users, passwords, "ops"))
	_, err = users.GetByName(ctx, "ops")
	assert.ErrorIs(t, err, storage.ErrNotFound)
	creds, err := access.LoadHtpasswd(f.Path)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)
	assert.False(t, f.Changed())

	assert.ErrorIs(t, f.Remove(ctx, users, passwords, "ops"), storage.ErrNotFound)
	assert.Error(t, f.Remove(ctx, users, passwords, ""))
}

// A file of one user makes that user an admin while storage holds none,
// as a storage whose admins all left would; with an admin in storage, a
// one-line file changes nobody.
func TestUsersFile_OneUserAdmin(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	users, passwords := newUsers(t)
	assert.NoError(t, users.Create(ctx, &model.User{Name: "ops", Role: model.RoleRunner}))
	f := app.NewUsersFile(htpasswdFile(t, "ops"))
	_, _, err := f.Sync(ctx, users, passwords)
	assert.NoError(t, err)
	assert.True(t, mustUser(t, users, "ops").Admin, "the one user, no admin")

	g := app.NewUsersFile(htpasswdFile(t, "ops", "dev"))
	_, _, err = g.Sync(ctx, users, passwords)
	assert.NoError(t, err)
	assert.False(t, mustUser(t, users, "dev").Admin)
	assert.Equal(t, model.RoleViewer, mustUser(t, users, "dev").Role)
}
