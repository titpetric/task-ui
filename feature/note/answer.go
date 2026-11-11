package note

import (
	"os"
	"strings"

	"github.com/titpetric/task-ui/app/model"
)

// replyBytes caps the reply file the board reads.
const replyBytes = 64 << 10

// replyFile makes the file a job writes its answer to; the reply on the
// note is that answer, and the run's output stays in the run.
func replyFile() (string, error) {
	answer, err := os.CreateTemp("", "note-reply-*.md")
	if err != nil {
		return "", err
	}
	return answer.Name(), answer.Close()
}

// readReply returns the reply file's contents, trimmed and cut to replyBytes, "" for a file it cannot read.
func readReply(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if len(data) > replyBytes {
		data = data[:replyBytes]
	}
	return strings.TrimSpace(string(data))
}

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
