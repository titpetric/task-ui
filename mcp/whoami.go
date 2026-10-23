package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/titpetric/task-ui/api"
)

// whoamiTool says what the token's user may do. An agent can't tell
// without trying, and a refused call points here (docs/api.md).
func whoamiTool() Tool {
	return Tool{
		Name: "whoami",
		Description: "Show who you are in task-ui and what you may do: your role, the permissions you hold, " +
			"each area of task-ui with what it is and what you may do there, and the folders you see.",
		InputSchema: object(map[string]Property{}),
		call:        whoami,
	}
}

func whoami(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var me api.MeView
	if err := c.get(ctx, "/me", &me); err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "You are %s, role %s.\nYou may: %s.\n\nAreas:\n", me.Name, me.Role, me.Summary)
	for _, area := range append(append([]api.AreaView{}, me.Sections...), me.Permissions...) {
		may := "denied"
		switch {
		case len(area.Rules) > 0:
			may = strings.Join(area.Rules, ", ")
		case area.Allowed:
			may = "allowed"
		}
		fmt.Fprintf(&b, "  %s (%s): %s. %s\n", area.Title, area.Key, may, area.About)
	}
	fmt.Fprintf(&b, "\nFolders you see (%d)", me.FolderCount)
	if me.FolderCount > len(me.Folders) {
		fmt.Fprintf(&b, ", the first %d", len(me.Folders))
	}
	fmt.Fprintf(&b, ": %s\n", strings.Join(me.Folders, ", "))
	if len(me.Keys) > 0 {
		fmt.Fprintf(&b, "\nSSH keys you may read with get_ssh_key: %s\n", strings.Join(me.Keys, ", "))
	}
	b.WriteString("\nResource " + contextURI + " describes every area.\n")
	return b.String(), nil
}
