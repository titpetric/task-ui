package claude

import (
	"github.com/titpetric/task-ui/config"
)

// Context is what Claude reads after its system prompt in every run: who it works for, then context.md.
// Without a user, Claude reads context.md alone.
func Context(user, role, permissions string) string {
	if user == "" {
		return config.Context()
	}
	line := "You work in task-ui for the user " + user
	if role != "" {
		line += ", role " + role
	}
	if permissions != "" {
		line += ", who may: " + permissions
	}
	return line + ". Do only what that user may do. What task-ui is follows.\n\n" + config.Context()
}
