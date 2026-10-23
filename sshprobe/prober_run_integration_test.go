//go:build integration

package sshprobe_test

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshprobe"
)

// runSSH is an ssh that runs the remote command locally for debian@a,
// and fails to connect anywhere else.
func runSSH(t *testing.T) string {
	t.Helper()
	script := filepath.Join(t.TempDir(), "ssh")
	body := "#!/bin/sh\nlogin=$(eval echo \\${$(($# - 1))})\nremote=$(eval echo \\${$#})\ncase \"$login\" in debian@a) exec sh -c \"$remote\";; esac\necho 'ssh: connect refused' >&2\nexit 255\n"
	writeScript(t, script, body)
	return script
}

// Argv is ssh with the login and the remote command last; on a
// terminal it gets -tt and the pagers are off.
func TestProber_Argv(t *testing.T) {
	t.Parallel()
	p := &sshprobe.Prober{Timeout: time.Second, Command: runSSH(t), Terminal: true}
	argv := p.Argv("debian@a", `printf 'pager=%s\n' "$PAGER"`)
	assert.Equal(t, p.Command, argv[0])
	assert.Equal(t, "-tt", argv[1])
	assert.Equal(t, "debian@a", argv[len(argv)-2])

	var out []byte
	var err error
	for range 5 {
		// A script written a moment ago can still be busy for exec.
		if out, err = exec.Command(argv[0], argv[1:]...).Output(); err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	assert.NoError(t, err)
	assert.Equal(t, "pager=cat\n", string(out))

	plain := (&sshprobe.Prober{Command: "ssh"}).Argv("debian@a", "true")
	assert.Equal(t, "-n", plain[1])
}
