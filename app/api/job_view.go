package api

import (
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/folder"
)

// JobView is a job as the jobs listing shows it: the job, whether the
// user asking may run it, and how it last went.
type JobView struct {
	*folder.Job
	CanRun  bool       `json:"can_run"`
	LastRun *model.Run `json:"last_run,omitempty"`
}
