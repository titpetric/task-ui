//go:build integration

package sshprobe_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/integration/ssh/sshprobe"
	"github.com/titpetric/task-ui/internal/assert"
)

// A host list comes back empty when the probe finds no login.
func TestProber_Logins(t *testing.T) {
	t.Parallel()
	p := &sshprobe.Prober{Users: []string{"debian"}, Command: closedCommand(t), Timeout: time.Second}
	assert.Empty(t, p.Logins(context.Background(), []string{"10.0.0.2"}, 200*time.Millisecond))
}

// A page that names a user tries that user before the instance's own.
func TestProber_LoginsAs(t *testing.T) {
	t.Parallel()
	p := &sshprobe.Prober{Users: []string{"debian"}, Command: openCommand(t), Timeout: time.Second}
	got := p.LoginsAs(context.Background(), []string{"10.0.0.2"}, []string{"ops", "root"}, time.Second)
	assert.Equal(t, "ops@10.0.0.2", got["10.0.0.2"])
}

// openCommand is an ssh client that accepts every login.
func openCommand(t *testing.T) string {
	t.Helper()
	return command(t, "#!/bin/sh\nexit 0\n")
}

// closedCommand is an ssh client that refuses every login.
func closedCommand(t *testing.T) string {
	t.Helper()
	return command(t, "#!/bin/sh\nexit 255\n")
}

// command writes a stub ssh client and returns its path.
func command(t *testing.T, script string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ssh")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
