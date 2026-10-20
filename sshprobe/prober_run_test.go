package sshprobe_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshprobe"
)

// runSSH is an ssh that runs the remote command locally for debian@a and
// debian@slow, and fails to connect anywhere else.
func runSSH(t *testing.T) string {
	t.Helper()
	script := filepath.Join(t.TempDir(), "ssh")
	body := "#!/bin/sh\nlogin=$(eval echo \\${$(($# - 1))})\nremote=$(eval echo \\${$#})\ncase \"$login\" in debian@a|debian@slow) exec sh -c \"$remote\";; esac\necho 'ssh: connect refused' >&2\nexit 255\n"
	assert.NoError(t, os.WriteFile(script, []byte(body), 0o755))
	return script
}

// Run runs on every host at once and reports each: the output, the exit
// code, a host it cannot reach, and one that runs past the time limit.
func TestProber_Run(t *testing.T) {
	t.Parallel()
	p := &sshprobe.Prober{Timeout: time.Second, Command: runSSH(t)}
	logins := map[string]string{"a": "debian@a", "down": "debian@down", "slow": "debian@slow"}
	command := `if [ "$(hostname)" = never ]; then :; fi; echo out; echo err >&2; sleep ${SLOW:-0}; exit 3`

	results := p.Run(context.Background(), []string{"a", "down", "nologin"}, logins, command, 5*time.Second)
	assert.Len(t, results, 3)
	a := results[0]
	assert.Equal(t, "debian@a", a.Login)
	assert.Equal(t, 3, a.ExitCode)
	assert.Contains(t, a.Output, "out")
	assert.Contains(t, a.Output, "err")
	assert.Equal(t, "", a.Error)
	assert.Equal(t, 255, results[1].ExitCode)
	assert.Contains(t, results[1].Error, "could not run")
	assert.Contains(t, results[2].Error, "no ssh login")

	slow := p.Run(context.Background(), []string{"slow"}, logins, "sleep 3", 300*time.Millisecond)
	assert.Contains(t, slow[0].Error, "ran past")

	big := p.Run(context.Background(), []string{"a"}, logins, "head -c 400000 /dev/zero | tr '\\\\0' x", 5*time.Second)
	assert.Equal(t, 256<<10, len(big[0].Output))
	assert.False(t, strings.Contains(big[0].Output, "\x00"))
}

// On a terminal, ssh gets -tt, the pagers are off, and the carriage
// returns of the terminal's line ends go.
func TestProber_RunTerminal(t *testing.T) {
	t.Parallel()
	p := &sshprobe.Prober{Timeout: time.Second, Command: runSSH(t), Terminal: true}
	logins := map[string]string{"a": "debian@a"}
	results := p.Run(context.Background(), []string{"a"}, logins, `printf 'pager=%s\r\n' "$PAGER"`, 5*time.Second)
	assert.Equal(t, "pager=cat\n", results[0].Output)
}
