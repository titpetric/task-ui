package agent

import (
	"os"
	"syscall"
)

// Shell replaces the process with Claude Code in a terminal, in the directory it starts in, with args and the workspace's home, TASK_UI_HOME.
// The pty is Claude's own, and the context goes into its system prompt.
// As a folder's @claude it starts in the folder, or in the directory inside it the run is of.
func (a *Agent) Shell(args []string) error {
	if !a.signedIn() {
		return errNoToken
	}
	path, err := a.claudePath()
	if err != nil {
		return err
	}
	wd, _ := os.Getwd()
	argv := append(append([]string{"claude", "--dangerously-skip-permissions", "--append-system-prompt", a.context()}, a.homeArgs(wd)...), args...)
	return syscall.Exec(path, argv, a.environ())
}
