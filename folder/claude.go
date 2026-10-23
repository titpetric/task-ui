package folder

import (
	"os"

	"github.com/titpetric/task-ui/model"
)

// ClaudeID is the ID of a folder's Claude job: Claude Code in a
// terminal in the folder's directory, or in the directory of a folder
// Sub returns.
const ClaudeID = model.ClaudeJob

// ClaudeJob returns the folder's Claude job, task-ui agent shell with a terminal in the folder's directory.
// A shell, as @shell is, for admins only, and no schedule or workflow step takes it.
func (f *Folder) ClaudeJob() *Job {
	if f == nil {
		return nil
	}
	job := &Job{
		Folder:  f.ID,
		Dir:     f.Rel,
		ID:      ClaudeID,
		Kind:    "shell",
		Command: []string{executable(), "agent", "shell"},
		Allow:   []string{model.RoleAdmin},
	}
	job.Name = "claude"
	job.Description = "Claude Code in a terminal in " + f.where() + "/"
	job.Interactive = true
	return job
}

// executable is the path of task-ui, whose agent subcommands the Claude
// jobs run.
func executable() string {
	exe, err := os.Executable()
	if err != nil {
		return "task-ui"
	}
	return exe
}

// where names the folder's directory in a job's description: the folder,
// or folder/dir for the folder of a directory inside it.
func (f *Folder) where() string {
	if f.Rel != "" {
		return f.ID + "/" + f.Rel
	}
	return f.ID
}
