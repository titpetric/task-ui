// Package runner runs jobs and keeps them running: a session per run,
// detached from any window, with a terminal emulator behind it so the
// screen is known at every moment, windows that attach and detach, a
// lock that allows one run per job across the instance, and the
// artifacts a run leaves behind.
package runner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/titpetric/task-ui/model"
)

// The errors a start or an attach can end in.
var (
	// ErrBusy is returned when the job is already running.
	ErrBusy = errors.New("already running")
	// ErrNotRunning is returned for input to a run that is not running.
	ErrNotRunning = errors.New("the run is not running")
	// ErrNotInteractive is returned for input to a job that takes none.
	ErrNotInteractive = errors.New("the job takes no input")
	// ErrNotFound is returned for a session that is not there.
	ErrNotFound = errors.New("no such run")
)

// Manager starts runs and holds their sessions.
type Manager struct {
	opts Options

	mu       sync.Mutex
	sessions map[string]*Session
	byKey    map[string]*Session
}

// New returns a manager. Zero durations get defaults.
func New(opts Options) *Manager {
	if opts.Grace <= 0 {
		opts.Grace = 5 * time.Second
	}
	if opts.KillTimeout <= 0 {
		opts.KillTimeout = 15 * time.Second
	}
	if opts.Keep <= 0 {
		opts.Keep = time.Hour
	}
	return &Manager{opts: opts, sessions: map[string]*Session{}, byKey: map[string]*Session{}}
}

// Get returns a session by run id.
func (m *Manager) Get(id string) (*Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[id]
	return session, ok
}

// Active returns the session running a job, by key.
func (m *Manager) Active(key string) (*Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.byKey[key]
	return session, ok
}

// List returns the sessions running, oldest first.
func (m *Manager) List() []*Session {
	m.mu.Lock()
	result := make([]*Session, 0, len(m.byKey))
	for _, session := range m.byKey {
		result = append(result, session)
	}
	m.mu.Unlock()
	sort.Slice(result, func(i, j int) bool { return result[i].Run.ID < result[j].Run.ID })
	return result
}

// Start starts a run. When the job is running already it returns the
// session in the way and ErrBusy, unless the request forces: then the
// run in the way is killed, waited for, and the new run takes its place.
func (m *Manager) Start(ctx context.Context, req Request) (*Session, error) {
	if req.Folder == nil || req.Job == nil {
		return nil, errors.New("a run needs a folder and a job")
	}
	run := m.newRun(req)
	if current, err := m.acquire(run.Key(), req.Force); err != nil {
		return current, err
	}
	if err := m.opts.Runs.Create(ctx, run); err != nil {
		m.release(run.Key(), nil)
		return nil, err
	}
	session := newSession(run, req.Folder, req.Job)
	m.mu.Lock()
	m.sessions[run.ID] = session
	m.byKey[run.Key()] = session
	m.mu.Unlock()
	go m.serve(session)
	return session, nil
}

func (m *Manager) newRun(req Request) *model.Run {
	run := &model.Run{
		Folder:   req.Folder.ID,
		Job:      req.Job.ID,
		TermCols: int64(req.Cols),
		TermRows: int64(req.Rows),
	}
	if req.User != nil {
		run.UserID, run.UserName = req.User.ID, req.User.Name
	}
	if req.Force {
		run.Forced = 1
	}
	if req.Scheduled {
		run.Scheduled = 1
	}
	if run.TermCols < 20 || run.TermCols > 500 {
		run.TermCols = 120
	}
	if run.TermRows < 5 || run.TermRows > 200 {
		run.TermRows = 30
	}
	return run
}

// acquire takes the lock of a key, killing the holder when forced.
func (m *Manager) acquire(key string, force bool) (*Session, error) {
	deadline := time.Now().Add(m.opts.KillTimeout)
	for {
		m.mu.Lock()
		current, busy := m.byKey[key]
		if !busy {
			m.byKey[key] = nil
			m.mu.Unlock()
			return nil, nil
		}
		m.mu.Unlock()
		if !force {
			return current, ErrBusy
		}
		current.Kill(m.opts.Grace)
		select {
		case <-current.Done():
		case <-time.After(time.Until(deadline)):
			return current, ErrBusy
		}
	}
}

// release frees the lock of a key, when the session holds it.
func (m *Manager) release(key string, session *Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.byKey[key] == session {
		delete(m.byKey, key)
	}
}

// forget drops a finished session from memory after a while.
func (m *Manager) forget(session *Session) {
	time.Sleep(m.opts.Keep)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sessions[session.ID()] == session {
		delete(m.sessions, session.ID())
	}
}

// Stop ends a run as stopped.
func (m *Manager) Stop(id string) error {
	session, ok := m.Get(id)
	if !ok {
		return ErrNotFound
	}
	session.Stop(m.opts.Grace)
	return nil
}

// Kill ends a run as killed.
func (m *Manager) Kill(id string) error {
	session, ok := m.Get(id)
	if !ok {
		return ErrNotFound
	}
	session.Kill(m.opts.Grace)
	return nil
}

// Shutdown kills every run and waits for the sessions to settle.
func (m *Manager) Shutdown(ctx context.Context) {
	for _, session := range m.List() {
		session.Kill(m.opts.Grace)
	}
	for _, session := range m.List() {
		select {
		case <-session.Done():
		case <-ctx.Done():
			return
		}
	}
}

// serve runs a session from start to finish: the recording, the pull, the
// env files, the process, the pump, the outcome, the artifacts.
func (m *Manager) serve(session *Session) {
	ctx := context.Background()
	run := session.Run

	rec, err := m.opts.Recordings.Open(run)
	if err == nil {
		session.addSink(rec)
		session.mu.Lock()
		run.CastPath = m.opts.Recordings.Relative(rec.Path())
		copied := *run
		session.mu.Unlock()
		_ = m.opts.Runs.Update(ctx, &copied)
	} else {
		slog.Warn("recording disabled for run", "run", run.ID, "err", err)
	}

	status := m.execute(ctx, session)
	if rec != nil {
		_ = rec.Close()
	}
	// The lock goes before the windows hear the outcome, so a window
	// that runs again on the status finds the job free.
	m.release(run.Key(), session)
	m.settle(ctx, session, status)
	go m.forget(session)
}

// settle writes the outcome down and tells the windows.
func (m *Manager) settle(ctx context.Context, session *Session, status Status) {
	run := session.Snapshot()
	run.State = status.State
	run.ExitCode = int64(status.ExitCode)
	if err := m.opts.Runs.Finish(ctx, &run); err != nil {
		slog.Warn("run outcome not recorded", "run", run.ID, "err", err)
	}
	session.mu.Lock()
	session.Run.FinishedAt = run.FinishedAt
	session.mu.Unlock()
	m.collectArtifacts(ctx, session)
	slog.Info("run finished", "folder", run.Folder, "job", run.Job, "user", run.UserName, "state", status.State, "exit", status.ExitCode)
	session.finish(status)
}

// execute does the run: a pull when the folder is a checkout, the env
// files, then the process, pumped until it exits.
func (m *Manager) execute(ctx context.Context, session *Session) Status {
	f, job, run := session.Folder, session.Job, session.Run
	if len(job.Command) == 0 {
		return failed(session, "the job has no command")
	}
	if m.opts.Git != nil && m.opts.Git.IsRepo(f.Dir) {
		fmt.Fprint(session, "\x1b[2m$ git pull --ff-only\x1b[0m\r\n")
		if err := m.opts.Git.Pull(ctx, f.Dir, crlfWriter{session}); err != nil {
			return failed(session, "git pull failed: "+err.Error())
		}
		fmt.Fprint(session, "\r\n")
	}
	if m.opts.Env != nil {
		if err := m.opts.Env.Generate(ctx, f); err != nil {
			return failed(session, "generating env files: "+err.Error())
		}
	}
	cmd, err := m.command(ctx, session)
	if err != nil {
		return failed(session, err.Error())
	}
	proc, err := StartProcess(cmd, int(run.TermCols), int(run.TermRows))
	if err != nil {
		return failed(session, "cannot start: "+err.Error())
	}
	session.setProcess(proc, m.opts.Grace)
	pump(proc, session)
	return session.outcome(proc.Wait())
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

// command builds the process: bash -l for the login environment, the
// job's argv passed as arguments so no name goes through the shell, the
// secrets and the run facts in the environment, the artifact directory
// ready.
func (m *Manager) command(ctx context.Context, session *Session) (*exec.Cmd, error) {
	f, job, run := session.Folder, session.Job, session.Run
	args := append([]string{"-l", "-c", `exec "$0" "$@"`}, job.Command...)
	cmd := exec.Command("bash", args...)
	cmd.Dir = f.Dir
	cmd.Env = os.Environ()
	if m.opts.Secrets != nil {
		secrets, err := m.opts.Secrets.For(ctx, f.ID)
		if err != nil {
			return nil, err
		}
		for _, key := range model.SortedKeys(secrets) {
			cmd.Env = append(cmd.Env, key+"="+secrets[key])
		}
	}
	cmd.Env = append(cmd.Env,
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"COLUMNS="+strconv.Itoa(int(run.TermCols)),
		"LINES="+strconv.Itoa(int(run.TermRows)),
		"TASK_UI=1",
		"TASK_UI_USER="+run.UserName,
		"TASK_UI_FOLDER="+f.ID,
		"TASK_UI_JOB="+job.ID,
		"TASK_UI_RUN_ID="+run.ID,
	)
	if dir := m.artifactDir(run); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		cmd.Env = append(cmd.Env, "TASK_UI_ARTIFACTS="+dir)
	}
	return cmd, nil
}

func (m *Manager) artifactDir(run *model.Run) string {
	if m.opts.ArtifactRoot == "" {
		return ""
	}
	return filepath.Join(m.opts.ArtifactRoot, run.ID)
}
