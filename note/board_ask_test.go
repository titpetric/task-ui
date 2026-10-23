package note_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/storage"
)

// AskArea puts a prompt of the Workflows page to Claude: a thread marked
// with the area, whose run gets the builder's instruction, and a reply
// in it keeps the area. AskArea refuses an area there is none of.
func TestBoard_AskArea(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()
	f.board.AgentCommand = []string{"sh", "-c", `printf '%s' "$TASK_UI_PROMPT_INSTRUCTIONS" > "$TASK_UI_PROMPT_REPLY"`}

	asked, err := f.board.AskArea(ctx, admin, model.AreaWorkflows, "proj", "build a weather workflow")
	assert.NoError(t, err)
	assert.Equal(t, model.AreaWorkflows, asked.Area)
	assert.Eventually(t, func() bool {
		got, _ := f.board.Notes.Get(ctx, asked.ID)
		return got.State == model.NoteDone
	}, 20*time.Second, 20*time.Millisecond)
	answer := replies(t, f, asked.ID)[0]
	assert.True(t, strings.HasPrefix(answer.Body, "The question describes a task-ui workflow"), answer.Body)
	assert.Equal(t, model.AreaWorkflows, answer.Area)
	threads, _ := f.board.Notes.List(ctx, storage.NoteFilter{Prompts: true, Dir: "proj"})
	assert.Len(t, threads, 0)

	_, err = f.board.AskArea(ctx, admin, "elsewhere", "proj", "x")
	assert.ErrorIs(t, err, note.ErrInvalid)
}
