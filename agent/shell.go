package agent

import (
	"os"
	"syscall"
)

// Shell replaces the process with Claude Code in a terminal, in the
// folders task-ui runs, TASK_UI_ROOT, with the arguments given: the pty is
// Claude's own. A development tool with you at the keyboard.
func (a *Agent) Shell(args []string) error {
	if a.getenv("CLAUDE_CODE_OAUTH_TOKEN") == "" {
		return errNoToken
	}
	path, err := a.claudePath()
	if err != nil {
		return err
	}
	if root := a.getenv("TASK_UI_ROOT"); root != "" {
		if err := os.Chdir(root); err != nil {
			return err
		}
	}
	argv := append([]string{"claude", "--dangerously-skip-permissions"}, args...)
	return syscall.Exec(path, argv, a.environ())
}
