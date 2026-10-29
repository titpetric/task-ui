package folder

import (
	"errors"

	"github.com/titpetric/task-ui/model"
)

// ScriptID is the ID of a folder's script job: the command a workflow's
// script step names, run by bash in the folder's directory.
const ScriptID = "@script"

// ErrScriptEmpty refuses a script step without a command.
var ErrScriptEmpty = errors.New("a script step needs a command")

// ScriptJob returns the folder's script job, which runs the step's command with bash -lc in the folder's directory.
// It exists for workflow steps alone: Job does not return it, and the shell permission gates it.
func (f *Folder) ScriptJob() *Job {
	if f == nil {
		return nil
	}
	job := &Job{
		Folder: f.ID,
		Dir:    f.Rel,
		ID:     ScriptID,
		Kind:   "script",
		Allow:  []string{model.RoleAdmin},
	}
	job.Name = "script"
	job.Description = "A command of a workflow step, run by bash in " + f.where() + "/"
	job.TakesArgs = true
	return job
}

// IsScript reports whether the job is a folder's script job, @script.
func (j *Job) IsScript() bool {
	return j != nil && j.ID == ScriptID && j.Kind == "script"
}
