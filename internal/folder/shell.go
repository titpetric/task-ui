package folder

import (
	"errors"

	"github.com/titpetric/task-ui/model"
)

// ShellID is the ID of a folder's shell job: an interactive login shell
// in the folder's directory, or in the directory of a folder Sub returns.
const ShellID = "@shell"

// ShellJob returns the folder's shell job, `bash -l` with a terminal in the folder's directory, for admins only.
// No allow rule of a folder or a job opens it to anyone else, and no schedule or workflow step takes it.
func (f *Folder) ShellJob() *Job {
	if f == nil {
		return nil
	}
	job := &Job{
		Folder:  f.ID,
		Dir:     f.Rel,
		ID:      ShellID,
		Kind:    "shell",
		Command: []string{"bash", "-l"},
		Allow:   []string{model.RoleAdmin},
	}
	job.Name = "shell"
	job.Description = "An interactive login shell in " + f.where() + "/"
	job.Interactive = true
	return job
}

// ErrShellUnattended refuses a shell job where no one is at the
// keyboard: a schedule or a workflow step.
var ErrShellUnattended = errors.New("an interactive shell needs someone at the keyboard, so a schedule or workflow step can't run it")

// IsShell reports whether the job is a folder's @shell or @claude, which the shell permission gates.
func (j *Job) IsShell() bool {
	return j != nil && (j.ID == ShellID || j.ID == ClaudeID) && j.Kind == "shell"
}
