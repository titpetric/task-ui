package api

import (
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/workflow"
)

// Options are what the handlers work with.
type Options struct {
	Users     *storage.UserStorage
	Projects  *storage.ProjectStorage
	Runs      *storage.RunStorage
	Artifacts *storage.ArtifactStorage
	Schedules *storage.ScheduleStorage
	Secrets   *secret.Store

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
	// KeyPolicies give the instance keys to roles' runs.
	KeyPolicies *storage.SSHKeyPolicyStorage
	Env         *envfile.Store
	Recordings  *recording.Store
	// Root is where projects are cloned to.
	Root string
}
