package runner

import (
	"time"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/storage"
)

// Options configure the manager.
type Options struct {
	Runs       *storage.RunStorage
	Artifacts  *storage.ArtifactStorage
	Secrets    *secret.Store
	Env        *envfile.Store
	Git        *git.Git
	// KeyDir is the instance SSH folder, and KeyPolicies which of its
	// keys each role's runs get through the run's ssh agent. Either nil
	// or empty gives no instance key to a run.
	KeyDir      string
	KeyPolicies *storage.SSHKeyPolicyStorage
	Recordings *recording.Store
	// ArtifactRoot is where collected artifact files go, one directory
	// per run.
	ArtifactRoot string
	// Grace is how long a stopped process gets before SIGKILL.
	Grace time.Duration
	// KillTimeout bounds how long Run force waits for the previous run.
	KillTimeout time.Duration
	// Keep is how long a finished session stays attachable in memory
	// before it is forgotten; the row and the recording remain.
	Keep time.Duration
	// Now is the clock the hours gate reads. Nil is time.Now.
	Now func() time.Time
}
