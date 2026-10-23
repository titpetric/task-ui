package sshprobe

import (
	"context"
)

// Argv returns the argv of ssh running a command on a login, as a
// run of @host:<host> starts it in a terminal of its own: batch mode,
// and with Terminal, -tt and the pagers off. The run's agent holds the
// instance keys.
func (p *Prober) Argv(login, remote string) []string {
	return p.terminal(context.Background(), login, remote).Args
}
