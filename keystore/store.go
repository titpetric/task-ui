package keystore

import (
	"context"

	"github.com/titpetric/task-ui/sshagent"
)

// Store is where the keys live. Local is the key folder; another store,
// such as an external service, keeps the same contract: nothing outside
// it holds a private key longer than a run.
type Store interface {
	// List returns every key.
	List(ctx context.Context) ([]Key, error)
	// Get returns a key by name.
	Get(ctx context.Context, name string) (*Key, error)
	// Generate makes a key; Import takes one in PEM or OpenSSH form.
	Generate(ctx context.Context, spec Spec) (*Key, error)
	Import(ctx context.Context, spec Spec, data []byte) (*Key, error)
	// Remove retires a key; Rotate writes new material for it, the old kept
	// to decrypt what it encrypted.
	Remove(ctx context.Context, name string) error
	Rotate(ctx context.Context, name string) (*Key, error)
	// Materialize serves SSH keys to a run in an ssh agent of its own;
	// closing the agent ends it.
	Materialize(ctx context.Context, names []string) (*sshagent.Agent, error)
}
