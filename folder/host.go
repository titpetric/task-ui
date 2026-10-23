package folder

import (
	"errors"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// HostPrefix starts the ID of a job of the root that runs a command on
// a host of the inventory: @host:dev1, or @host:* for every host with
// an ssh login.
const HostPrefix = model.HostPrefix

// ErrHostBare refuses the job @host, which names no host.
var ErrHostBare = errors.New("a command on a host is the job @host:<name>, a host of the inventory, or @host:* for every host with an ssh login")

// HostJob returns the root's job that runs a command on host over ssh, or on every host with an ssh login for *.
// Nil for any other folder, a directory inside the root, or no host.
func (f *Folder) HostJob(host string) *Job {
	if f == nil || f.ID != model.Default || f.Rel != "" || host == "" || strings.ContainsAny(host, " \t\r\n/") {
		return nil
	}
	job := &Job{
		Folder: f.ID,
		ID:     model.HostJob(host),
		Kind:   "host",
	}
	job.Name = job.ID
	job.Description = "A command on " + host + " over ssh"
	if host == "*" {
		job.Description = "A command on every host with an ssh login, a run a host"
	}
	job.TakesArgs = true
	return job
}

// IsHost reports whether the job is a host job of the root, @host:<host>.
func (j *Job) IsHost() bool {
	return j != nil && j.Kind == "host" && strings.HasPrefix(j.ID, HostPrefix)
}

// HostTarget returns the host a host job runs on, * for every host, ""
// for any other job.
func (j *Job) HostTarget() string {
	if !j.IsHost() {
		return ""
	}
	return strings.TrimPrefix(j.ID, HostPrefix)
}
