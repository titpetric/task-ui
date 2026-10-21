package runner

import (
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// Request is what starts a run.
type Request struct {
	Folder *folder.Folder
	Job    *folder.Job
	User   *model.User
	Force  bool
	// Reason says why the request forces a run out of hours; the run keeps it.
	Reason    string
	Scheduled bool
	Cols      int
	Rows      int
	// Env adds NAME=value pairs to the run's environment, after the
	// secrets and before the TASK_UI_ variables, which win.
	Env []string
	// Workspace is the directory of the workspace the run belongs to, the
	// root for the default one. Empty takes the workspace of the folder's
	// first segment, or the default workspace for a folder in none.
	Workspace string
	// WorkspaceOf is a folder ID whose workspace the run takes when
	// Workspace is empty: for an internal job run for a folder, such as a
	// prompt in it. Empty takes the workspace of the run's own folder.
	WorkspaceOf string
	// Dir is the directory the job runs in, absolute: a directory inside
	// the folder, which the caller holds to the folder's rules, for a job
	// of Folder.Sub. Empty runs in the folder's own directory.
	Dir string
	// Args follow the job's argv after --, as argv and never through a
	// shell: only for a job that takes args.
	Args []string
}
