package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"

	"github.com/titpetric/task-ui/model"
)

// Process is a command running on a pseudo terminal, in its own session,
// so stopping it stops everything it started.
type Process struct {
	cmd *exec.Cmd
	tty *os.File

	stopOnce sync.Once
	exited   chan struct{}
}

// StartProcess runs a command on a pty of the given size.
func StartProcess(cmd *exec.Cmd, cols, rows int) (*Process, error) {
	tty, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
	if err != nil {
		return nil, err
	}
	return &Process{cmd: cmd, tty: tty, exited: make(chan struct{})}, nil
}

// Read reads the process output.
func (p *Process) Read(b []byte) (int, error) {
	return p.tty.Read(b)
}

// Write sends input to the process.
func (p *Process) Write(b []byte) (int, error) {
	return p.tty.Write(b)
}

// Close closes the terminal side.
func (p *Process) Close() error {
	return p.tty.Close()
}

// PID returns the process id.
func (p *Process) PID() int {
	return p.cmd.Process.Pid
}

// Resize changes the terminal size. It ignores a size that is not positive.
func (p *Process) Resize(cols, rows int) {
	if cols > 0 && rows > 0 {
		_ = pty.Setsize(p.tty, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
	}
}

// Stop sends SIGHUP and SIGTERM to the process group, and SIGKILL when it has not exited after grace.
// Stopping a process that exited does nothing.
func (p *Process) Stop(grace time.Duration) {
	select {
	case <-p.exited:
		return
	default:
	}
	p.stopOnce.Do(func() {
		// pty.Start puts the child in a new session, its pid is the
		// process group id.
		pid := p.cmd.Process.Pid
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

// Wait waits for the process and returns its exit code: the status, or
// 128 plus the signal that ended it.
func (p *Process) Wait() int {
	err := p.cmd.Wait()
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

// failed writes why in red and returns the failed status.
func failed(session *Session, message string) Status {
	fmt.Fprintf(session, "\x1b[31m%s\x1b[0m\r\n", message)
	return Status{State: model.StateFailed, ExitCode: 1, Message: message}
}

// pump copies the process output into the session until it ends.
func pump(proc *Process, session *Session) {
	buf := make([]byte, 32*1024)
	for {
		n, err := proc.Read(buf)
		if n > 0 {
			session.write(buf[:n])
		}
		if err != nil {
			_ = proc.Close()
			return
		}
	}
}
