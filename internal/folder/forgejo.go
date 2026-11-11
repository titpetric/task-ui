package folder

import (
	"github.com/titpetric/task-ui/model"
)

// ForgejoSyncID is the ID of the root's Forgejo sync job: task-ui
// forgejo-sync, which makes a workspace of every organisation of the
// Forgejo integration and a project of every repository.
const ForgejoSyncID = "@forgejo-sync"

// ForgejoSyncJob returns the root's Forgejo sync job, task-ui forgejo-sync, or nil for any other folder or a directory in it.
// Admins run it from the Integrations page, by hand, on a schedule or in a workflow step; --dry-run prints what it would do.
func (f *Folder) ForgejoSyncJob() *Job {
	if f == nil || f.ID != model.Default || f.Rel != "" {
		return nil
	}
	job := &Job{
		Folder:  f.ID,
		ID:      ForgejoSyncID,
		Kind:    "forgejo-sync",
		Command: []string{executable(), "forgejo-sync"},
		Allow:   []string{model.RoleAdmin},
	}
	job.Name = "forgejo-sync"
	job.Description = "Sync the Forgejo organisations: workspaces, homes, projects and push hooks"
	job.TakesArgs = true
	return job
}
