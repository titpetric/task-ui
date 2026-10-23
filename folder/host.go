package folder

import (
	"github.com/titpetric/task-ui/model"
)

// HostID is the ID of the root's host job: a command on a host of the
// inventory, over the ssh login task-ui has to it.
const HostID = model.HostJob

// HostJob returns the host job of the root, _, nil for any other folder
// or a directory inside the root. Its run names the host as its target
// and the command as its args; the hosts permission says who runs it.
// Its command is the ssh client's, which the run's starter fills in for
// the host, so the job alone has none. Unlike @shell it needs no one at
// the keyboard: a schedule or a workflow step takes it.
func (f *Folder) HostJob() *Job {
	if f == nil || f.ID != model.Default || f.Rel != "" {
		return nil
	}
	job := &Job{
		Folder: f.ID,
		ID:     HostID,
		Kind:   "host",
	}
	job.Name = "host"
	job.Description = "A command on a host of the inventory, over ssh"
	job.TakesArgs = true
	return job
}

// IsHost reports whether the job is the root's host job, @host.
func (j *Job) IsHost() bool {
	return j != nil && j.ID == HostID && j.Kind == "host"
}
