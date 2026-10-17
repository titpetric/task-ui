package folder

import (
	"github.com/titpetric/task-ui/detector"
)

// Job is one runnable target of a folder.
type Job struct {
	detector.Target

	Folder string `json:"folder"`
	// ID is the job's name within the folder: the target name when the
	// folder follows one convention, kind:name when it follows several,
	// so `make build` and `npm run build` stay apart.
	ID string `json:"id"`
	// Kind and Source name the convention and file the job came from.
	Kind   string `json:"kind"`
	Source string `json:"source"`

	// Command is the full argv: the source command and the target's args.
	Command []string `json:"command"`

	// Allow narrows who may run this job, on top of the folder Allow,
	// from the configuration.
	Allow []string `json:"allow,omitempty"`
	// Hours says when this job may run, over the folder's hours, from
	// the configuration.
	Hours string `json:"hours,omitempty"`
}

// Key is the identity used for the run lock and the history: one run per
// job per folder, across the instance.
func (j *Job) Key() string {
	return j.Folder + "/" + j.ID
}
