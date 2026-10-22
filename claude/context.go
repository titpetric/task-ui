package claude

import (
	"github.com/titpetric/task-ui/config"
)

// Context is what Claude reads after its own system prompt in every run
// task-ui starts: who it works for, with the role and a line of what the
// user may do, then `config/context.md`, what task-ui is. No user leaves
// out the first line.
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
