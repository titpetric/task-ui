//go:build integration

package sshprobe_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshprobe"
)

// fakeSSH is an ssh that signs in as debian to a, as root to b, slowly to
// c, and to nothing else; it records its arguments.
func fakeSSH(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	log := filepath.Join(dir, "args")
	script := filepath.Join(dir, "ssh")
	body := "#!/bin/sh\necho \"$@\" >> " + log + "\nfor last; do :; done\ntarget=$(eval echo \\${$(($# - 1))})\ncase \"$target\" in debian@a|root@b) exit 0;; *@c) sleep 2; exit 0;; esac\nexit 255\n"
	writeScript(t, script, body)
	return script, log
}

// writeScript writes an executable script and waits until it runs. A
// child that a parallel test forks while the file is open for writing
// holds the open file until it execs, and running the script then fails
// with "text file busy". Once one run starts, no child holds it.
func writeScript(t *testing.T, path, body string) {
	t.Helper()
	assert.NoError(t, os.WriteFile(path, []byte(body), 0o755))
	for i := 0; i < 100; i++ {
		if err := exec.Command(path).Run(); !errors.Is(err, syscall.ETXTBSY) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestProber_Logins(t *testing.T) {
	t.Parallel()
	ssh, log := fakeSSH(t)
	keys := &agentKeys{t: t}
	p := &sshprobe.Prober{Users: []string{"debian", "root"}, Keys: keys, KnownHosts: "/dev/null", Timeout: 5 * time.Second, TTL: time.Hour, Command: ssh}

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
	assert.Contains(t, string(args), "IdentityAgent="+filepath.Join(keys.dir, "task-ui-agent-"), "the keys come from an agent")
	assert.NotContains(t, string(args), "-i ")
	assert.Contains(t, string(args), "BatchMode=yes")
}
