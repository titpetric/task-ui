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
	// Reason says why a run is forced out of hours; it is kept on the run.
	Reason    string
	Scheduled bool
	Cols      int
	Rows      int
}
