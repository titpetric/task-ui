package workflow

import (
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// Step is a workflow step resolved to its folder and job, or to the
// workflow it runs.
type Step struct {
	Folder *folder.Folder
	Job    *folder.Job
	// Workflow is the workflow the step runs, nil for a job.
	Workflow *model.Workflow
	// Disabled is a step the run skips, resolved to nothing.
	Disabled bool
	// Asks is a prompt or a summary step, which asks Claude in Place, a
	// folder and a directory in it or internal for the root; Summary is a
	// summary step.
	Asks    bool
	Summary bool
	Place   string
	// Args follow the job's argv after --, for a job that takes args;
	// for a step of @host:<host> or @script they are the shell command.
	Args []string
	// Group is the index of the project the step runs in, for a workflow
	// that runs over a workspace: a failure skips the rest of its
	// project's steps alone. 0 for every step of any other workflow.
	Group int
}
