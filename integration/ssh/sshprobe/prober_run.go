package sshprobe

import (
	"context"
)

// Argv returns the argv of the ssh that a run of @host:<host> starts, as terminal builds it.
func (p *Prober) Argv(login, remote string) []string {
	return p.terminal(context.Background(), login, remote).Args
}
