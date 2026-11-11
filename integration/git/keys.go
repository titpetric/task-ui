package git

import (
	"context"

	"github.com/titpetric/task-ui/sshagent"
)

// Keys hand git its key for a clone or a pull: an agent holding it and
// the path of its public half, or a nil agent for none, and the
// known_hosts file.
type Keys interface {
	GitAgent(ctx context.Context) (*sshagent.Agent, string, error)
	KnownHosts() string
}
