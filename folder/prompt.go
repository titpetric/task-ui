package folder

import (
	"github.com/titpetric/task-ui/model"
)

// PromptID is the ID of a folder's prompt job: Claude Code on a prompt
// in the folder's directory, or in the directory of a folder Sub
// returns, as the user who asks.
const PromptID = model.PromptJob

// PromptJob returns the folder's prompt job, task-ui agent prompt in the folder's directory.
// A prompt starts it with its thread from the prompt routes, which hold it to prompt/edit; Job does not resolve it.
func (f *Folder) PromptJob() *Job {
	if f == nil {
		return nil
	}
	job := &Job{
		Folder:  f.ID,
		Dir:     f.Rel,
		ID:      PromptID,
		Kind:    "prompt",
		Command: []string{executable(), "agent", "prompt"},
	}
	job.Name = "claude-prompt"
	job.Description = "Claude Code on a prompt in " + f.where() + "/"
	return job
}

// IsPrompt reports whether the job is a folder's prompt job.
func (j *Job) IsPrompt() bool {
	return j != nil && j.ID == PromptID && j.Kind == "prompt"
}
