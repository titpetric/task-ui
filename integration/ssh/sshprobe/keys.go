package sshprobe

import (
	"context"

	"github.com/titpetric/task-ui/sshagent"
)

// Keys serve the keys a probe signs in with, in an ssh agent the probe
// closes; a nil agent for none.
type Keys interface {
	LoginAgent(ctx context.Context) (*sshagent.Agent, error)
}
