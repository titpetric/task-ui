package note

import (
	"github.com/titpetric/task-ui/app/model"
)

// Conversation is a thread's messages and the work of each run.
type Conversation struct {
	Thread   *model.Note       `json:"thread"`
	Agent    string            `json:"agent"`
	Agents   []string          `json:"agents"`
	Messages []model.Note      `json:"messages"`
	Runs     []ConversationRun `json:"runs"`
}
