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
	"github.com/titpetric/task-ui/hostrun"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/retention"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
)

// Options are what the handlers work with.
type Options struct {
	Users    *storage.UserStorage
	Projects *storage.ProjectStorage
	// Workspaces are the directories the projects may clone inside,
	Workspaces *storage.WorkspaceStorage
	Runs       *storage.RunStorage
	Artifacts  *storage.ArtifactStorage
	// ArtifactStore is the tree of the files the runs leave.
	ArtifactStore *artifactstore.Store
	Schedules     *storage.ScheduleStorage
	Secrets       *secret.Store
	// SecretReads log which run got which secret. Nil lists no reads.
	SecretReads *storage.SecretReadStorage
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
	// Steps is the catalogue a workflow step picks its folder from: every
	// folder with a job, the ones the workspace and project settings hide
	// too, as the workflow editor lists them. A run start and a pin
	// resolve their folder through it as well. Nil picks from Catalogue.
	Steps folder.Catalogue
	Git   *git.Git
	// Keys are the key folder: every key of the instance and its
	// workspaces.
	Keys       *keystore.Local
	Env        *envfile.Store
	Recordings *recording.Store
	// Cleaner holds the Data retention rules; nil turns their route off.
	Cleaner *retention.Cleaner
	// Root is the folder that holds cloned projects.
	Root string
	// Board is the notes board and the chat; nil turns them off.
	Board *note.Board
	// Integrations keep the integrations' settings, and Hosts reads the
	// host inventory through the Beszel one.
	Integrations *integration.Store
	Hosts        *beszel.Inventory
	// HostRuns start the runs of @host:<host>, a command on hosts of the
	// inventory. Nil runs on no host.
	HostRuns *hostrun.Starter
	// AddUser puts a new user in the password file; nil leaves users to
	// the file and task-ui passwd.
	AddUser func(ctx context.Context, name, password string) error
	// RemoveUser takes a user out of the password file and storage at
	// once; nil leaves removing to the file.
	RemoveUser func(ctx context.Context, name string) error
	// Tokens keep the tokens clients sign in with; a user's MCP token
	// is one. Nil turns MCP access from the settings off.
	Tokens *storage.UserAuthTokenStorage
	// Now is the clock a token's expiry starts from; nil is the real
	// time.
	Now func() time.Time
	// Pins keep what each user pins to the sidebar; nil turns /pins off.
	Pins PinStore
	// Summaries keep Claude's summary of each run that has one.
	Summaries *storage.RunSummaryStorage
	// Summarizer writes the summary of a run on the hosts with Claude.
	Summarizer *claude.Summarizer
	// Version is the task-ui an export names in its manifest.
	Version string
	// ExportWorkFactor is the scrypt work factor of an export's
	// passphrase, log2 of N; zero is age's default.
	ExportWorkFactor int
	// Tools keep the tools the Instance page found in the PATH, checked
	// on demand; nil turns the check routes off.
	Tools *storage.ToolStorage
}
