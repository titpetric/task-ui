package folder

import (
	"errors"

	"github.com/titpetric/task-ui/model"
)

// ShellID is the ID of a folder's shell job: an interactive login shell
// in the folder's directory, or in the directory of a folder Sub returns.
const ShellID = "@shell"

// ShellJob returns the folder's shell job, `bash -l` with a terminal in
// the folder's directory, for admins only: no allow rule of a folder or a
// job opens it to anyone else. No schedule or workflow step takes it.
func (f *Folder) ShellJob() *Job {
	if f == nil {
		return nil
	}
	where := f.ID
	if f.Rel != "" {
		where += "/" + f.Rel
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
	job.Description = "An interactive login shell in " + where + "/"
	job.Interactive = true
	return job
}

// ErrShellUnattended refuses the shell job where no one is at the
// keyboard: a schedule or a workflow step.
var ErrShellUnattended = errors.New("an interactive shell runs with an admin at the keyboard: no schedule or workflow step takes it")

// IsShell reports whether the job is a folder's shell job.
func (j *Job) IsShell() bool {
	return j != nil && j.ID == ShellID && j.Kind == "shell"
}
