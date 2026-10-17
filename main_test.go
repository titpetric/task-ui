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
	"github.com/titpetric/task-ui/model"
)

func TestHash(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	assert.NoError(t, hash([]string{"secret"}, strings.NewReader(""), &out))
	assert.True(t, access.CheckPassword(&model.User{Password: strings.TrimSpace(out.String())}, "secret"))

	out.Reset()
	assert.NoError(t, hash(nil, strings.NewReader("from stdin\n"), &out))
	assert.True(t, access.CheckPassword(&model.User{Password: strings.TrimSpace(out.String())}, "from stdin"))

	assert.Error(t, hash(nil, strings.NewReader(""), &out))
}

func TestRun(t *testing.T) {
	var out bytes.Buffer
	root := t.TempDir()
	t.Setenv("TASK_UI_ROOT", root)
	t.Setenv("TASK_UI_DATA", filepath.Join(root, "data"))
	t.Setenv("TASK_UI_LISTEN", "127.0.0.1:0")
	t.Setenv("TASK_UI_SSH_DIR", filepath.Join(root, "ssh"))

	assert.NoError(t, run(context.Background(), []string{"hash", "pw"}, strings.NewReader(""), &out))
	assert.True(t, access.Hashed(strings.TrimSpace(out.String())))

	out.Reset()
	assert.NoError(t, run(context.Background(), []string{"-v"}, strings.NewReader(""), &out))
	assert.Contains(t, out.String(), "task-ui dev")

	assert.Error(t, run(context.Background(), []string{"--nope"}, strings.NewReader(""), &out))

	// passwd adds a user to the password file, and resets one; a bad file
	// stops the server before it listens.
	t.Setenv("TASK_UI_HTPASSWD", filepath.Join(root, ".htpasswd"))
	assert.Error(t, run(context.Background(), []string{"passwd"}, strings.NewReader(""), &out))
	out.Reset()
	assert.NoError(t, run(context.Background(), []string{"passwd", "alice", "secret"}, strings.NewReader(""), &out))
	assert.Contains(t, out.String(), "alice: password set")
	assert.NoError(t, run(context.Background(), []string{"passwd", "alice"}, strings.NewReader("reset\n"), &out))
	creds, err := access.LoadHtpasswd(filepath.Join(root, ".htpasswd"))
	assert.NoError(t, err)
	assert.Len(t, creds, 1)
	assert.True(t, access.CheckPassword(&model.User{Password: creds[0].Hash}, "reset"))
	assert.Error(t, run(context.Background(), []string{"passwd", "bob"}, strings.NewReader(""), &out))

	assert.NoError(t, os.WriteFile(filepath.Join(root, ".htpasswd"), []byte("broken\n"), 0o600))
	assert.Error(t, run(context.Background(), nil, strings.NewReader(""), &out))
	assert.NoError(t, os.Remove(filepath.Join(root, ".htpasswd")))
	t.Setenv("TASK_UI_HTPASSWD", "")

	// The server runs until the context ends.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	assert.NoError(t, run(ctx, nil, strings.NewReader(""), &out))
	assert.FileExists(t, filepath.Join(root, "data", "task-ui.db"))
}
