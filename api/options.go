package api

import (
	"context"

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
)

// Options are what the handlers work with.
type Options struct {
	Users     *storage.UserStorage
	Projects  *storage.ProjectStorage
	Runs      *storage.RunStorage
	Artifacts *storage.ArtifactStorage
	// ArtifactStore is the tree of the files the runs leave.
	ArtifactStore *artifactstore.Store
	Schedules     *storage.ScheduleStorage
	Secrets       *secret.Store

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
	// Summarizer writes the summary of a run on the hosts with Claude.
	Summarizer *claude.Summarizer
}
