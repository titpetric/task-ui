package note

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// TestConversation asserts a conversation carries the thread and the work of each run.
func TestConversation(t *testing.T) {
	conversation := &Conversation{
		Thread:   &model.Note{ID: "n", Body: "question"},
		Agent:    "claude",
		Agents:   []string{"claude"},
		Messages: []model.Note{{ID: "n", Body: "question"}},
		Runs:     []ConversationRun{{Question: "n", Run: "r"}},
	}
	data, err := json.Marshal(conversation)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"agent":"claude"`)
	assert.Contains(t, string(data), `"runs":[{"question":"n","run":"r","answer":"","events":null}]`)
}
