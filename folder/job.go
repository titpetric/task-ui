package folder

import (
	"sort"
	"strings"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/model"
)

// Job is one runnable target of a folder.
type Job struct {
	detector.Target

	Folder string `json:"folder"`
	// Dir is the directory inside the folder the job runs in, a slash
	// path: empty for a job of the folder itself, set for a job of a
	// directory inside it (see Folder.Sub).
	Dir string `json:"dir,omitempty"`
	// ID is the job's name within the folder: the target name when the
	// folder follows one convention, kind:name when it follows more than one,
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
// job per folder and directory, across the instance.
func (j *Job) Key() string {
	return model.RunKey(j.Folder, j.Dir, j.ID)
}

// ArgsCommand is what a run of a job that takes args starts, as a page
// shows it before the arguments: the argv of its exec: step, which the
// arguments follow, or the job's own argv and --. Empty for a job that
// takes none.
func (j *Job) ArgsCommand() string {
	if j == nil || !j.TakesArgs {
		return ""
	}
	if len(j.Exec) > 0 {
		return strings.Join(j.Exec, " ")
	}
	return strings.Join(j.Command, " ") + " --"
}

// IsDefault reports whether the job is a default target: an ID of
// default, after the kind or skill in front of it when it has one.
func (j *Job) IsDefault() bool {
	if j == nil {
		return false
	}
	id := strings.TrimPrefix(j.ID, SkillPrefix)
	return id == "default" || strings.HasSuffix(id, ":default")
}

// DefaultFirst moves the default jobs to the front of a list, in
// place, the rest keeping their order. Load and Skills order a folder's
// jobs with it, so default is the first job everywhere.
func DefaultFirst(jobs []*Job) []*Job {
	sort.SliceStable(jobs, func(a, b int) bool {
		return jobs[a].IsDefault() && !jobs[b].IsDefault()
	})
	return jobs
}
