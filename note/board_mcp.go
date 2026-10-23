package note

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/titpetric/task-ui/model"
)

// mcpServer names task-ui's MCP server in a run's MCP configuration.
const mcpServer = "task"

// mcpEnv returns the environment that connects a Claude run the board
// starts to task-ui's MCP server as the user: a run token, minted now, in an MCP configuration file of a 0700
// directory of its own, which TASK_UI_MCP_CONFIG names. The returned
// cleanup revokes the token and removes the directory, once the run
// ended. Without RunTokens or an MCP URL, or when the file is not made,
// the run goes without, and the cleanup does nothing.
func (b *Board) mcpEnv(user *model.User) ([]string, func()) {
	none := func() {}
	if b.RunTokens == nil || b.MCPURL == nil || user == nil {
		return nil, none
	}
	url := b.MCPURL()
	if url == "" {
		return nil, none
	}
	token, revoke, err := b.RunTokens.Issue(user.ID)
	if err != nil {
		slog.Warn("run token not made", "user", user.Name, "err", err)
		return nil, none
	}
	dir, err := os.MkdirTemp("", "task-ui-mcp-*")
	if err != nil {
		revoke()
		slog.Warn("MCP configuration not made", "user", user.Name, "err", err)
		return nil, none
	}
	cleanup := func() {
		revoke()
		os.RemoveAll(dir)
	}
	config := map[string]any{"mcpServers": map[string]any{mcpServer: map[string]any{
		"type": "http", "url": url, "headers": map[string]string{"Authorization": "Bearer " + token},
	}}}
	data, err := json.Marshal(config)
	path := filepath.Join(dir, "mcp.json")
	if err == nil {
		err = os.WriteFile(path, data, 0o600)
	}
	if err != nil {
		cleanup()
		slog.Warn("MCP configuration not written", "user", user.Name, "err", err)
		return nil, none
	}
	return []string{"TASK_UI_MCP_CONFIG=" + path}, cleanup
}
