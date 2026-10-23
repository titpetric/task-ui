package folder

import (
	"github.com/titpetric/task-ui/model"
)

// CleanupID is the ID of the root's cleanup job: task-ui cleanup, which
// applies the Data retention settings.
const CleanupID = "@cleanup"

// CleanupJob returns the root's cleanup job, task-ui cleanup, or nil for any other folder or a directory in it.
// Admins run it by hand, on the daily schedule task-ui seeds, or in a workflow step; --dry-run prints what would go.
func (f *Folder) CleanupJob() *Job {
	if f == nil || f.ID != model.Default || f.Rel != "" {
		return nil
	}
	job := &Job{
		Folder:  f.ID,
		ID:      CleanupID,
		Kind:    "cleanup",
		Command: []string{executable(), "cleanup"},
		Allow:   []string{model.RoleAdmin},
	}
	job.Name = "cleanup"
	job.Description = "Apply the Data retention settings: remove what they let go"
	job.TakesArgs = true
	return job
}
