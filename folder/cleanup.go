package folder

import (
	"github.com/titpetric/task-ui/model"
)

// CleanupID is the ID of the root's cleanup job: task-ui cleanup, which
// applies the Data retention settings.
const CleanupID = "@cleanup"

// CleanupJob returns the cleanup job of the root, _, nil for any other
// folder or a directory inside the root: task-ui cleanup, which removes
// the recordings, the artifact versions and the secret reads the Data
// retention settings let go. Admins alone run it, by hand, on the daily
// schedule task-ui seeds, or in a workflow step; --dry-run as its args
// prints what would go.
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
