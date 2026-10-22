package agent

import (
	"syscall"
)

// Shell replaces the process with Claude Code in a terminal, in the
// directory it starts in, with the arguments given: the pty is Claude's
// own. As a folder's @claude it starts in the folder, or in the
// directory inside it the run is of. A development tool with you at the
// keyboard, with the context in its system prompt.
func (a *Agent) Shell(args []string) error {
	if a.getenv("CLAUDE_CODE_OAUTH_TOKEN") == "" {
		return errNoToken
	}
	path, err := a.claudePath()
	if err != nil {
		return err
	}
	argv := append([]string{"claude", "--dangerously-skip-permissions", "--append-system-prompt", a.context()}, args...)
	return syscall.Exec(path, argv, a.environ())
}
