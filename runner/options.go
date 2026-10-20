package runner

import (
	"time"

	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/storage"
)

// Options configure the manager.
type Options struct {
	Runs      *storage.RunStorage
	Artifacts *storage.ArtifactStorage
	Secrets   *secret.Store
	Env       *envfile.Store
	// Integrations add to a run what an integration has for its folder,
	// such as the Claude Code token. Nil adds nothing.
	Integrations Integrations
	Git          *git.Git
	// KeyDir is the instance SSH folder, and KeyPolicies which of its
	// keys each role's runs get through the run's ssh agent. With either
	// nil or empty, a run gets no instance key.
	KeyDir      string
	KeyPolicies *storage.SSHKeyPolicyStorage
	Recordings  *recording.Store
	// ArtifactRoot is where a run writes its artifacts, one directory per
	// run, while it runs.
	ArtifactRoot string
	// ArtifactStore takes a run's artifacts when it ends, into the one tree
	// of every run's files. Nil keeps them in the run's directory.
	ArtifactStore *artifactstore.Store
	// Grace is how long a stopped process gets before SIGKILL.
	Grace time.Duration
	// KillTimeout bounds how long Run force waits for the previous run.
	KillTimeout time.Duration
	// Keep is how long a finished session stays attachable in memory
	// before the manager drops it; the row and the recording remain.
	Keep time.Duration
	// Now is the clock the hours gate reads. Nil is time.Now.
	Now func() time.Time
}
