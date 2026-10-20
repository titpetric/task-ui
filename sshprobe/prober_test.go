package sshprobe_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshprobe"
)

// fakeSSH is an ssh that signs in as debian to a, as root to b, slowly to
// c, and to nothing else; it records the keys it was handed.
func fakeSSH(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	log := filepath.Join(dir, "args")
	script := filepath.Join(dir, "ssh")
	body := "#!/bin/sh\necho \"$@\" >> " + log + "\nfor last; do :; done\ntarget=$(eval echo \\${$(($# - 1))})\ncase \"$target\" in debian@a|root@b) exit 0;; *@c) sleep 2; exit 0;; esac\nexit 255\n"
	assert.NoError(t, os.WriteFile(script, []byte(body), 0o755))
	return script, log
}

func TestProber_Logins(t *testing.T) {
	t.Parallel()
	ssh, log := fakeSSH(t)
	keys := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(keys, "id_ed25519"), []byte("k"), 0o600))
	assert.NoError(t, os.WriteFile(filepath.Join(keys, "id_ed25519.pub"), []byte("k"), 0o644))
	p := &sshprobe.Prober{Users: []string{"debian", "root"}, KeyDirs: []string{keys}, KnownHosts: "/dev/null", Timeout: 5 * time.Second, TTL: time.Hour, Command: ssh}

	logins := p.Logins(context.Background(), []string{"a", "b", "z", "c", ""}, 500*time.Millisecond)
	assert.Equal(t, "debian@a", logins["a"])
	assert.Equal(t, "root@b", logins["b"])
	assert.Equal(t, "", logins["z"])
	assert.Equal(t, "", logins["c"], "still being tried")

	assert.Eventually(t, func() bool {
		return p.Logins(context.Background(), []string{"c"}, 0)["c"] == "debian@c"
	}, 5*time.Second, 50*time.Millisecond)

	args, err := os.ReadFile(log)
	assert.NoError(t, err)
	assert.Contains(t, string(args), "-i "+filepath.Join(keys, "id_ed25519")+" ")
	assert.NotContains(t, string(args), "id_ed25519.pub")
	assert.Contains(t, string(args), "BatchMode=yes")
}
