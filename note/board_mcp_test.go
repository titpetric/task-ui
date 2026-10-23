package note_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/storage"
)

// A prompt's run gets TASK_UI_MCP_CONFIG, an MCP configuration with a
// run token of the asker that signs in while the run goes; once it
// ended, the token signs in nobody and the board removed the directory.
// Without an MCP URL the run goes without.
func TestBoard_MCPConfig(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()
	tokens := access.NewRunTokens()
	f.board.RunTokens = tokens
	f.board.MCPURL = func() string { return "http://127.0.0.1:1/api/mcp" }
	f.board.AgentCommand = []string{"sh", "-c", `{ echo "$TASK_UI_MCP_CONFIG"; cat "$TASK_UI_MCP_CONFIG"; } > "$TASK_UI_PROMPT_REPLY"`}

	asked, err := f.board.Ask(ctx, admin, "", "proj", "list the jobs")
	assert.NoError(t, err)
	var reply string
	assert.Eventually(t, func() bool {
		threads, _ := f.board.Notes.List(ctx, storage.NoteFilter{Prompts: true, Dir: "proj"})
		if len(threads) != 1 || threads[0].State == "open" || threads[0].State == "working" {
			return false
		}
		reply = replies(t, f, asked.ID)[0].Body
		return true
	}, 20*time.Second, 20*time.Millisecond)
	path, config, _ := strings.Cut(reply, "\n")
	assert.Equal(t, "mcp.json", filepath.Base(path))
	assert.Contains(t, config, `"url":"http://127.0.0.1:1/api/mcp"`)
	_, token, _ := strings.Cut(config, `"Authorization":"Bearer `)
	token, _, _ = strings.Cut(token, `"`)
	// The cleanup follows the answer, a moment after it.
	assert.Eventually(t, func() bool {
		_, live := tokens.User(token)
		_, err := os.Stat(filepath.Dir(path))
		return !live && os.IsNotExist(err)
	}, 10*time.Second, 20*time.Millisecond)

	f.board.MCPURL = func() string { return "" }
	_, err = f.board.Ask(ctx, admin, "", "proj/images", "list the jobs")
	assert.NoError(t, err)
}
