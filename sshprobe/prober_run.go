package sshprobe

import (
	"context"
)

// Argv returns the argv of ssh running a command on a login, as a
// run of @host:<host> starts it in a terminal of its own: the instance keys,
// batch mode, and with Terminal, -tt and the pagers off.
func (p *Prober) Argv(login, remote string) []string {
	return p.terminal(context.Background(), login, remote).Args
}
