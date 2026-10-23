package note_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
)

// A prompt's run with task-ui's tools reads its source, the thread to
// reply in; a reply it posts with RunReply is Claude's, in the thread,
// and its final message is then no second reply. A run that posts none
// replies with its final message, the fallback. On a board note, no
// prompt's, RunReply comments.
func TestBoard_RunReply(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()
	out := t.TempDir()
	f.board.RunTokens = access.NewRunTokens()
	f.board.MCPURL = func() string { return "http://127.0.0.1:1/api/mcp" }
	f.board.AgentCommand = []string{"sh", "-c", `printf '%s' "$TASK_UI_PROMPT_SOURCE" > ` + out + `/source; ` +
		`i=0; while [ ! -f ` + out + `/go ] && [ $i -lt 600 ]; do sleep 0.05; i=$((i+1)); done; rm -f ` + out + `/go; echo fallback > "$TASK_UI_PROMPT_REPLY"`}
	source := func() string {
		var text []byte
		assert.Eventually(t, func() bool {
			var err error
			text, err = os.ReadFile(filepath.Join(out, "source"))
			return err == nil && len(text) > 0
		}, 10*time.Second, 20*time.Millisecond)
		assert.NoError(t, os.Remove(filepath.Join(out, "source")))
		return string(text)
	}
	release := func() { assert.NoError(t, os.WriteFile(filepath.Join(out, "go"), nil, 0o644)) }

	thread, err := f.board.Ask(ctx, admin, "", "proj/images", "what builds here")
	assert.NoError(t, err)
	text := source()
	assert.Contains(t, text, "The prompt is the thread of the task-ui note "+thread.ID+", in proj/images, /f/_/proj/images#note-"+thread.ID+".")
	assert.Contains(t, text, "note_comment with the id "+thread.ID)
	reply, err := f.board.RunReply(ctx, admin, thread.ID, "the images", "done")
	assert.NoError(t, err)
	assert.Equal(t, "claude", reply.UserName)
	release()
	assert.Equal(t, model.NoteDone, f.settled(t, thread.ID).State)
	all := replies(t, f, thread.ID)
	assert.Len(t, all, 1)
	assert.Equal(t, "the images", all[0].Body)

	asked, err := f.board.Ask(ctx, admin, thread.ID, "", "which first")
	assert.NoError(t, err)
	assert.Contains(t, source(), ", its question the reply "+asked.ID+", in proj/images")
	release()
	assert.Eventually(t, func() bool { return len(replies(t, f, thread.ID)) == 3 }, 10*time.Second, 20*time.Millisecond)
	assert.Equal(t, model.NoteDone, f.settled(t, thread.ID).State)
	assert.Equal(t, "fallback", replies(t, f, thread.ID)[2].Body)

	_, err = f.board.RunReply(ctx, admin, thread.ID, " ", "")
	assert.ErrorIs(t, err, note.ErrInvalid)
	board, err := f.board.Post(ctx, admin, "", "a note on the board")
	assert.NoError(t, err)
	comment, err := f.board.RunReply(ctx, admin, board.ID, "seen", "")
	assert.NoError(t, err)
	assert.Equal(t, "admin", comment.UserName)
}
