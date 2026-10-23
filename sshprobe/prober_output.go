package sshprobe

import (
	"context"
)

// Output runs a command on a login in batch mode, with no terminal, and
// returns what it printed on stdout.
func (p *Prober) Output(ctx context.Context, login, remote string) ([]byte, error) {
	return p.ssh(ctx, login, remote).Output()
}
