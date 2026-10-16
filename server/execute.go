package server

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/history"
	"github.com/titpetric/task-ui/run"
)

// execute runs the job command on a pty and pumps it to the socket and
// the recording. It returns how the run ended; an error is a run that
// never started.
func (svc *Server) execute(conn *socket, f *folder.Folder, job *folder.Job, active *run.Run, cols, rows int) (string, int, error) {
	if len(job.Command) == 0 {
		return "", 0, errors.New("job has no command")
	}

	// bash -l gives the login environment (PATH, profile), and passing the
	// command as arguments keeps the target name out of shell parsing.
	args := append([]string{"-l", "-c", `exec "$0" "$@"`}, job.Command...)
	cmd := exec.Command("bash", args...)
	cmd.Dir = f.Dir

	// The env files the sources read are generated with the secrets first,
	// and the secrets go into the process environment as well, so a source
	// without env files gets them too.
	if err := svc.env.Generate(f); err != nil {
		return "", 0, fmt.Errorf("generating env files: %w", err)
	}
	cmd.Env = os.Environ()
	for key, value := range svc.secrets.For(f.ID) {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	cmd.Env = append(cmd.Env,
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"COLUMNS="+strconv.Itoa(cols),
		"LINES="+strconv.Itoa(rows),
		"TASK_UI=1",
		"TASK_UI_USER="+active.User,
		"TASK_UI_FOLDER="+f.ID,
		"TASK_UI_PROJECT="+f.ID,
		"TASK_UI_JOB="+job.ID,
		"TASK_UI_RUN_ID="+active.ID,
	)

	conn.status(control{State: history.StateRunning, ID: active.ID, User: active.User, Started: active.Started.UTC().Format(time.RFC3339)})

	var recorder *history.Recorder
	if svc.config.History {
		var err error
		recorder, err = svc.history.Start(history.Record{
			ID:      active.ID,
			Folder:  f.ID,
			Job:     job.ID,
			User:    active.User,
			Started: active.Started,
			State:   history.StateRunning,
			Forced:  active.Forced,
		}, cols, rows)
		if err != nil {
			slog.Warn("history disabled for run", "err", err)
			recorder = nil
		}
	}
	finish := func(state string, exitCode int) {
		if recorder != nil {
			if err := recorder.Finish(state, exitCode); err != nil {
				slog.Warn("history write failed", "err", err)
			}
		}
	}
	// term is what the user sees: the socket, and the recording when on.
	var term io.Writer = conn
	if recorder != nil {
		term = io.MultiWriter(conn, recorder)
	}

	// A folder that is a checkout is brought up to date first, so the run
	// uses the latest version of the project. A pull that fails fails the
	// run: running stale code quietly is worse than not running.
	if svc.git.IsRepo(f.Dir) {
		fmt.Fprintf(term, "\x1b[2m$ git pull --ff-only\x1b[0m\r\n")
		if err := svc.git.Pull(conn.request().Context(), f.Dir, crlfWriter{term}); err != nil {
			fmt.Fprintf(term, "\x1b[31mgit pull failed: %s\x1b[0m\r\n", err)
			finish(history.StateFailed, 1)
			return history.StateFailed, 1, nil
		}
		fmt.Fprint(term, "\r\n")
	}

	proc, err := run.Start(cmd, cols, rows)
	if err != nil {
		finish(history.StateFailed, 1)
		return "", 0, err
	}
	defer proc.TTY.Close()

	var stopped bool
	var stopMu sync.Mutex
	stop := func(byUser bool) {
		stopMu.Lock()
		stopped = stopped || byUser
		stopMu.Unlock()
		proc.Stop(5 * time.Second)
	}
	active.SetStop(func() { stop(false) })

	// Client to process: input for interactive jobs, resize and stop for
	// all. The read fails when the window closes, which stops the run.
	go func() {
		for {
			msg, err := conn.receive()
			if err != nil {
				stop(true)
				return
			}
			switch msg.Type {
			case "resize":
				proc.Resize(msg.Cols, msg.Rows)
			case "stop":
				fmt.Fprint(conn, "\r\n\x1b[33mStopping...\x1b[0m\r\n")
				stop(true)
			case "input":
				if job.Interactive {
					_, _ = proc.TTY.Write([]byte(msg.Message))
				}
			}
		}
	}()

	// Process to client and recording.
	buf := make([]byte, 32*1024)
	for {
		n, err := proc.TTY.Read(buf)
		if n > 0 {
			_, _ = term.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}

	exitCode := proc.Wait()

	stopMu.Lock()
	byUser := stopped
	stopMu.Unlock()

	state := history.StateOK
	switch {
	case active.Killed():
		state = history.StateKilled
	case byUser:
		state = history.StateStopped
	case exitCode != 0:
		state = history.StateFailed
	}
	finish(state, exitCode)
	return state, exitCode, nil
}

// crlfWriter turns the newlines of a plain command into the line ends a
// terminal expects, which a pty would otherwise do.
type crlfWriter struct {
	w io.Writer
}

func (c crlfWriter) Write(p []byte) (int, error) {
	_, err := c.w.Write([]byte(strings.ReplaceAll(string(p), "\n", "\r\n")))
	return len(p), err
}
