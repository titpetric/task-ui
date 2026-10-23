package runner

import (
	"time"

	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workspace"
)

// Options configure the manager.
type Options struct {
	Runs      *storage.RunStorage
	Artifacts *storage.ArtifactStorage
	// Summaries are Claude's summaries of the runs; a run started again
	// under its ID drops its own. Nil keeps none.
	Summaries *storage.RunSummaryStorage
	Secrets   *secret.Store
	// SecretReads log which run got which secret. Nil logs none.
	SecretReads *storage.SecretReadStorage
	Env         *envfile.Store
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
	// Root is where the folders are, and Workspaces the checkouts under it
	// whose folder is a namespace: a run of a folder inside one names the
	// workspace in its environment. Nil Workspaces knows none.
	Root       string
	Workspaces *storage.WorkspaceStorage
	// WorkspaceKeys are the keys of the workspaces, and KeyGrants whose
	// runs get each of a workspace's SSH keys: a run gets its
	// workspace's keys beside the instance ones. Nil WorkspaceKeys adds
	// none; with nil KeyGrants every run gets every key.
	WorkspaceKeys *workspace.Keyring
	KeyGrants     *storage.WorkspaceKeyGrantStorage
	// Users say what the user of a run may do, which the run reads in
	// TASK_UI_PERMISSIONS, so Claude in it knows. Nil leaves it out.
	Users *storage.UserStorage
}
