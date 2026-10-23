//go:build integration

package sshprobe_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshprobe"
)

// Output runs a command on a login, with no terminal, and returns its
// stdout; a login that does not sign in is an error.
func TestProber_Output(t *testing.T) {
	t.Parallel()
	p := &sshprobe.Prober{Timeout: time.Second, Command: runSSH(t), Terminal: true}
	var out []byte
	var err error
	for range 5 {
		if out, err = p.Output(context.Background(), "debian@a", `echo "tty=$PAGER"`); err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	assert.NoError(t, err)
	assert.Equal(t, "tty=\n", string(out))
	_, err = p.Output(context.Background(), "root@b", "true")
	assert.Error(t, err)
}
