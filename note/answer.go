package note

import (
	"github.com/titpetric/task-ui/model"
)

// Answer returns the reply of a thread that a run of Claude's wrote:
// the last reply naming the run that is not at work, nil until the run
// answered.
func Answer(replies []model.Note, runID string) *model.Note {
	if runID == "" {
		return nil
	}
	for i := len(replies) - 1; i >= 0; i-- {
		if reply := &replies[i]; reply.RunID == runID && reply.State != model.NoteWorking && reply.Body != "" {
			return reply
		}
	}
	return nil
}
