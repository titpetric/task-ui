package runner

import (
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// Request is what starts a run.
type Request struct {
	Folder    *folder.Folder
	Job       *folder.Job
	User      *model.User
	Force     bool
	Scheduled bool
	Cols      int
	Rows      int
}
