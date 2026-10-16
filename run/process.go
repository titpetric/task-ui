package run

import (
	"errors"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
)

// Process is a command running on a pseudo terminal, in its own session, so
// stopping it stops everything it started.
type Process struct {
	Cmd *exec.Cmd
	TTY *os.File

	stopOnce sync.Once
	exited   chan struct{}
}

// Start runs a command on a pty of the given size.
func Start(cmd *exec.Cmd, cols, rows int) (*Process, error) {
	tty, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
	if err != nil {
		return nil, err
	}
	return &Process{Cmd: cmd, TTY: tty, exited: make(chan struct{})}, nil
}

// Resize changes the terminal size.
func (p *Process) Resize(cols, rows int) {
	if cols > 0 && rows > 0 {
		_ = pty.Setsize(p.TTY, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
	}
}

// Stop sends SIGTERM to the process group, and SIGKILL when it has not
// exited after the grace period.
func (p *Process) Stop(grace time.Duration) {
	select {
	case <-p.exited:
		return
	default:
	}
	p.stopOnce.Do(func() {
		pid := p.Cmd.Process.Pid
		// pty.Start puts the child in a new session, its pid is the
		// process group id.
		_ = syscall.Kill(-pid, syscall.SIGHUP)
		_ = syscall.Kill(-pid, syscall.SIGTERM)
		go func() {
			select {
			case <-p.exited:
			case <-time.After(grace):
				_ = syscall.Kill(-pid, syscall.SIGKILL)
			}
		}()
	})
}

// Wait waits for the process and returns its exit code.
func (p *Process) Wait() int {
	err := p.Cmd.Wait()
	close(p.exited)
	if err == nil {
		return 0
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		if status, ok := exit.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			return 128 + int(status.Signal())
		}
		return exit.ExitCode()
	}
	return -1
}
