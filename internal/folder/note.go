package folder

import (
	"github.com/titpetric/task-ui/model"
)

// NoteID is the ID of the Notes agent's job: Claude Code on an
// instruction from the Notes board, a job of the root, _.
const NoteID = model.NoteJob

// NoteJob returns the folder's Notes agent job, task-ui agent note in the folder's directory.
// The board starts it on a note as the agent's user, as _/@claude-note; Job does not resolve it.
func (f *Folder) NoteJob() *Job {
	if f == nil {
		return nil
	}
	job := &Job{
		Folder:  f.ID,
		Dir:     f.Rel,
		ID:      NoteID,
		Kind:    "note",
		Command: []string{executable(), "agent", "note"},
	}
	job.Name = "claude-note"
	job.Description = "Claude Code on an instruction from the Notes board, in " + f.where() + "/"
	return job
}

// IsNote reports whether the job is the Notes agent's job.
func (j *Job) IsNote() bool {
	return j != nil && j.ID == NoteID && j.Kind == "note"
}
