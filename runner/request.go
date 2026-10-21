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
	// Workspace is the directory of the workspace the run belongs to. Empty
	// takes the workspace of the folder's first segment, when there is one.
	Workspace string
}
