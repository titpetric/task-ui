package agent

import (
	"os"
	"syscall"
)

// Shell replaces the process with Claude Code in the working directory, with args, the run context in --append-system-prompt and TASK_UI_HOME in --add-dir.
// Claude owns the pty. It returns errNoToken when no token is set.
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
