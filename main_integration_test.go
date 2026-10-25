//go:build integration

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestRun(t *testing.T) {
	var out bytes.Buffer
	root := t.TempDir()
	t.Setenv("HOME", filepath.Join(root, "serverhome"))
	t.Setenv("TASK_UI_WORKSPACE", root)
	t.Setenv("TASK_UI_DATA", filepath.Join(root, "data"))
	t.Setenv("TASK_UI_LISTEN", "127.0.0.1:0")
	t.Setenv("TASK_UI_SSH_DIR", filepath.Join(root, "ssh"))

	assert.NoError(t, run(context.Background(), []string{"hash", "pw"}, strings.NewReader(""), &out))
	assert.True(t, access.Hashed(strings.TrimSpace(out.String())))

	out.Reset()
	assert.NoError(t, run(context.Background(), []string{"-v"}, strings.NewReader(""), &out))
	assert.Contains(t, out.String(), "task-ui dev")

	assert.Error(t, run(context.Background(), []string{"--nope"}, strings.NewReader(""), &out))

	// cleanup applies the rules to the data of the configuration.
	out.Reset()
	assert.NoError(t, run(context.Background(), []string{"cleanup", "--", "--dry-run"}, strings.NewReader(""), &out))
	assert.Contains(t, out.String(), "a dry run: nothing changes")
	assert.Contains(t, out.String(), "Run recordings (keep for 31 days, no size limit): would remove 0 recordings, 0 B")
	assert.Error(t, run(context.Background(), []string{"cleanup", "--all"}, strings.NewReader(""), &out))

	// The channel needs a token file and an instance; it ends with stdin.
	assert.Error(t, run(context.Background(), []string{"channel"}, strings.NewReader(""), &out))
	tokenFile := filepath.Join(root, "token")
	assert.NoError(t, os.WriteFile(tokenFile, []byte("tui_x\n"), 0o600))
	assert.Error(t, run(context.Background(), []string{"channel", "--token-file", tokenFile, "--url", ""}, strings.NewReader(""), &out))
	assert.NoError(t, run(context.Background(), []string{"channel", "--token-file", tokenFile, "--url", "http://127.0.0.1:1"}, strings.NewReader(""), &out))

	// agent reaches the host commands, before the configuration.
	assert.ErrorContains(t, run(context.Background(), []string{"agent"}, strings.NewReader(""), &out), "usage: task-ui agent")
	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "")
	assert.ErrorContains(t, run(context.Background(), []string{"agent", "note", "--root", "/nowhere"}, strings.NewReader(""), &out), "TASK_UI_NOTE: this job runs from a note")

	// passwd adds a user to the password file, <data>/.htpasswd, and
	// resets one; a bad file stops the server before it listens.
	htpasswd := filepath.Join(root, "data", ".htpasswd")
	assert.Error(t, run(context.Background(), []string{"passwd"}, strings.NewReader(""), &out))
	out.Reset()
	assert.NoError(t, run(context.Background(), []string{"passwd", "alice", "secret"}, strings.NewReader(""), &out))
	assert.Contains(t, out.String(), "alice: password set")
	assert.NoError(t, run(context.Background(), []string{"passwd", "alice"}, strings.NewReader("reset\n"), &out))
	creds, err := access.LoadHtpasswd(htpasswd)
	assert.NoError(t, err)
	assert.Len(t, creds, 1)
	assert.True(t, access.CheckHash(creds[0].Hash, "reset"))
	assert.Error(t, run(context.Background(), []string{"passwd", "bob"}, strings.NewReader(""), &out))

	assert.NoError(t, os.WriteFile(htpasswd, []byte("broken\n"), 0o600))
	assert.Error(t, run(context.Background(), nil, strings.NewReader(""), &out))
	assert.NoError(t, os.Remove(htpasswd))

	// The variables of the older layout stop it with what replaced them.
	t.Setenv("TASK_UI_ROOT", root)
	assert.ErrorContains(t, run(context.Background(), nil, strings.NewReader(""), &out), "TASK_UI_ROOT is not read")
	assert.NoError(t, os.Unsetenv("TASK_UI_ROOT"))

	// The server runs until the context ends.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	assert.NoError(t, run(ctx, nil, strings.NewReader(""), &out))
	assert.FileExists(t, filepath.Join(root, "data", "task-ui.db"))
}

// email, backup and token change what a user from the password file
// signs in with, in the database the server reads.
func TestRun_EmailAndBackup(t *testing.T) {
	var out bytes.Buffer
	root := t.TempDir()
	t.Setenv("TASK_UI_WORKSPACE", root)
	t.Setenv("TASK_UI_DATA", filepath.Join(root, "data"))
	t.Setenv("TASK_UI_SSH_DIR", filepath.Join(root, "keys"))
	ctx := context.Background()

	// Not a user until the server has read the file.
	assert.ErrorContains(t, run(ctx, []string{"email", "alice"}, strings.NewReader(""), &out), "no user alice")
	assert.NoError(t, run(ctx, []string{"passwd", "alice", "secret"}, strings.NewReader(""), &out))
	served, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	t.Setenv("TASK_UI_LISTEN", "127.0.0.1:0")
	assert.NoError(t, run(served, nil, strings.NewReader(""), &out))

	out.Reset()
	assert.NoError(t, run(ctx, []string{"email", "alice", "add", "alice@example.com", "a@example.com"}, strings.NewReader(""), &out))
	assert.Equal(t, "alice@example.com\na@example.com\n", out.String())
	out.Reset()
	assert.NoError(t, run(ctx, []string{"email", "alice", "rm", "a@example.com"}, strings.NewReader(""), &out))
	assert.Equal(t, "alice@example.com\n", out.String())
	assert.Error(t, run(ctx, []string{"email", "alice", "add", "nope"}, strings.NewReader(""), &out))
	assert.Error(t, run(ctx, []string{"email"}, strings.NewReader(""), &out))
	assert.Error(t, run(ctx, []string{"email", "alice", "add"}, strings.NewReader(""), &out))

	out.Reset()
	assert.NoError(t, run(ctx, []string{"backup", "alice", "3"}, strings.NewReader(""), &out))
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	assert.Len(t, lines, 4)
	assert.Contains(t, lines[0], "each signs in once")
	assert.NoError(t, run(ctx, []string{"backup", "alice"}, strings.NewReader(""), &out))
	assert.Error(t, run(ctx, []string{"backup", "alice", "x"}, strings.NewReader(""), &out))
	assert.Error(t, run(ctx, []string{"backup", "alice", "0"}, strings.NewReader(""), &out))
	assert.Error(t, run(ctx, []string{"backup"}, strings.NewReader(""), &out))
	assert.Error(t, run(ctx, []string{"backup", "bob"}, strings.NewReader(""), &out))

	// The command prints a token once; the list names it without it.
	out.Reset()
	assert.NoError(t, run(ctx, []string{"token", "alice", "mcp"}, strings.NewReader(""), &out))
	secret := strings.TrimSpace(out.String())
	assert.True(t, strings.HasPrefix(secret, "tui_"))
	out.Reset()
	assert.NoError(t, run(ctx, []string{"token", "alice"}, strings.NewReader(""), &out))
	assert.Equal(t, "mcp\tnever used\n", out.String())
	assert.NoError(t, run(ctx, []string{"token", "alice", "rm", "mcp"}, strings.NewReader(""), &out))
	out.Reset()
	assert.NoError(t, run(ctx, []string{"token", "alice"}, strings.NewReader(""), &out))
	assert.Equal(t, "", out.String())
	assert.Error(t, run(ctx, []string{"token"}, strings.NewReader(""), &out))
	assert.Error(t, run(ctx, []string{"token", "alice", "rm"}, strings.NewReader(""), &out))
	assert.Error(t, run(ctx, []string{"token", "alice", "add", "x"}, strings.NewReader(""), &out))
	assert.Error(t, run(ctx, []string{"token", "bob", "x"}, strings.NewReader(""), &out))
}

// user sets the admin flag, the role and the MCP flags of the users in
// the database, without the server; the last admin keeps the flag
// unless forced.
func TestRun_User(t *testing.T) {
	var out bytes.Buffer
	root := t.TempDir()
	t.Setenv("TASK_UI_WORKSPACE", root)
	t.Setenv("TASK_UI_DATA", filepath.Join(root, "data"))
	t.Setenv("TASK_UI_SSH_DIR", filepath.Join(root, "keys"))
	ctx := context.Background()
	assert.NoError(t, run(ctx, []string{"passwd", "alice", "secret"}, strings.NewReader(""), &out))
	assert.NoError(t, run(ctx, []string{"passwd", "bob", "secret"}, strings.NewReader(""), &out))
	served, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	t.Setenv("TASK_UI_LISTEN", "127.0.0.1:0")
	assert.NoError(t, run(served, nil, strings.NewReader(""), &out))

	out.Reset()
	assert.NoError(t, run(ctx, []string{"user", "list"}, strings.NewReader(""), &out))
	assert.Equal(t, "alice\tviewer\tadmin\nbob\tviewer\tadmin\n", out.String(), "the first users of an empty database are admins")

	out.Reset()
	assert.NoError(t, run(ctx, []string{"user", "admin", "alice", "off"}, strings.NewReader(""), &out))
	assert.Equal(t, "alice\tviewer\t\n", out.String())
	assert.ErrorContains(t, run(ctx, []string{"user", "admin", "bob", "off"}, strings.NewReader(""), &out), "the last admin")
	assert.NoError(t, run(ctx, []string{"user", "admin", "bob", "off", "--force"}, strings.NewReader(""), &out))
	assert.NoError(t, run(ctx, []string{"user", "admin", "bob", "on"}, strings.NewReader(""), &out))

	out.Reset()
	assert.NoError(t, run(ctx, []string{"user", "role", "alice", "runner"}, strings.NewReader(""), &out))
	assert.NoError(t, run(ctx, []string{"user", "agent", "alice", "on"}, strings.NewReader(""), &out))
	assert.Equal(t, "alice\trunner\t\nalice\trunner\tmcp,agent\n", out.String(), "agent reply turns MCP on")
	out.Reset()
	assert.NoError(t, run(ctx, []string{"user", "mcp", "alice", "off"}, strings.NewReader(""), &out))
	assert.Equal(t, "alice\trunner\t\n", out.String(), "MCP off ends agent reply")

	assert.ErrorContains(t, run(ctx, []string{"user", "role", "alice", "admin"}, strings.NewReader(""), &out), "runner or viewer")
	assert.ErrorContains(t, run(ctx, []string{"user", "mcp", "alice", "maybe"}, strings.NewReader(""), &out), "usage")
	assert.ErrorContains(t, run(ctx, []string{"user", "admin", "carol", "on"}, strings.NewReader(""), &out), "no user carol")
	assert.ErrorContains(t, run(ctx, []string{"user"}, strings.NewReader(""), &out), "usage")
}
