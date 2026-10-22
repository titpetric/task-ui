package api

import (
	"context"
	"time"

	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/sshprobe"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
	"github.com/titpetric/task-ui/workspace"
)

// Options are what the handlers work with.
type Options struct {
	Users    *storage.UserStorage
	Projects *storage.ProjectStorage
	// Workspaces are the checkouts the projects may clone inside,
	// WorkspaceKeys their keys, and KeyGrants whose runs get each of a
	// workspace's SSH keys.
	Workspaces    *storage.WorkspaceStorage
	WorkspaceKeys *workspace.Keyring
	KeyGrants     *storage.WorkspaceKeyGrantStorage
	Runs          *storage.RunStorage
	Artifacts     *storage.ArtifactStorage
	// ArtifactStore is the tree of the files the runs leave.
	ArtifactStore *artifactstore.Store
	Schedules     *storage.ScheduleStorage
	Secrets       *secret.Store
	// FolderRefs find the rows that name a folder, which a move leaves.
	FolderRefs *storage.FolderRefStorage

	Workflows    *storage.WorkflowStorage
	WorkflowRuns *storage.WorkflowRunStorage
	// Flows runs the workflows.
	Flows *workflow.Runner
	// Events is the bus /sse/{topic} streams from.
	Events *event.Bus

	Runner    *runner.Manager
	Catalogue folder.Catalogue
	Git       *git.Git
	Keys      *sshkey.Store
	// KeyPolicies name the roles whose runs get each instance key.
	KeyPolicies *storage.SSHKeyPolicyStorage
	Env         *envfile.Store
	Recordings  *recording.Store
	// Root is the folder that holds cloned projects.
	Root string
	// Board is the notes board and the chat; nil turns them off.
	Board *note.Board
	// Integrations keep the integrations' settings, and Hosts reads the
	// host inventory through the Beszel one.
	Integrations *integration.Store
	Hosts        *beszel.Inventory
	// SSH runs a command on the hosts over the logins it finds.
	SSH *sshprobe.Prober
	// AddUser puts a new user in the password file; nil leaves users to
	// the file and task-ui passwd.
	AddUser func(ctx context.Context, name, password string) error
	// Tokens keep the tokens clients sign in with; a user's MCP token
	// is one. Nil turns MCP access from the settings off.
	Tokens *storage.UserAuthTokenStorage
	// Now is the clock a token's expiry starts from; nil is the real
	// time.
	Now func() time.Time
	// Summaries keep Claude's summary of each run that has one.
	Summaries *storage.RunSummaryStorage
	// Summarizer writes the summary of a run on the hosts with Claude.
	Summarizer *claude.Summarizer
	// Version is the task-ui an export names in its manifest.
	Version string
	// ExportWorkFactor is the scrypt work factor of an export's
	// passphrase, log2 of N; zero is age's default.
	ExportWorkFactor int
}
