// Package runner runs jobs and keeps them running: a session per run,
// detached from any window, with a terminal emulator behind it that
// holds the screen at every moment, windows that attach and detach, a
// lock that allows one run per job across the instance, and the
// artifacts a run leaves behind. A prompt's runs and the Notes agent's
// are each a Claude session of their own and go side by side.
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
	"strings"
	"sync"
	"time"

	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/model"
)

// The errors a start or an attach can end in.
var (
	// ErrBusy reports a job that is already running.
	ErrBusy = errors.New("already running")
	// ErrNotRunning refuses input to a run that is not running.
	ErrNotRunning = errors.New("the run has already ended")
	// ErrNotInteractive refuses input to a job that takes none.
	ErrNotInteractive = errors.New("the job takes no input")
	// ErrNotFound reports a session that does not exist.
	ErrNotFound = errors.New("no such run")
	// ErrMoving refuses a run of a folder whose project is moving, and a
	// second move of it.
	ErrMoving = errors.New("the project is moving")
	// ErrNoArgs refuses arguments for a job that takes none.
	ErrNoArgs = errors.New("the job takes no arguments")
	// ErrQueueFull refuses a run that would wait behind MaxQueued runs of
	// its job.
	ErrQueueFull = fmt.Errorf("%d runs of the job wait already; start it again once one of them started", MaxQueued)
)

// MaxQueued is how many runs of a job wait at most for the run going.
const MaxQueued = 10

// Manager starts runs and holds their sessions.
type Manager struct {
	opts Options
	// Events gets the runs topic told of each run as it starts and as
	// it ends: a `run` event with the run row. Nil tells nobody.
	Events *event.Bus

	mu       sync.Mutex
	sessions map[string]*Session
	byKey    map[string]*Session
	// queued are the sessions that wait for the session of their key, in
	// the order they came; closing starts none of them any more.
	queued  map[string][]*Session
	closing bool
	// pulls holds one lock per checkout, so its runs pull one at a time.
	pulls map[string]*sync.Mutex
	// frozen are the folders that are moving: no run of them or of a
	// folder inside them starts.
	frozen map[string]bool
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
	return &Manager{opts: opts, sessions: map[string]*Session{}, byKey: map[string]*Session{}, queued: map[string][]*Session{}}
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
// session in the way and ErrBusy, unless the request forces: then Start
// kills the run in the way, waits for it, and the new run takes its
// place; or unless the request queues: then the new run waits, queued,
// and starts once the runs before it ended.
func (m *Manager) Start(ctx context.Context, req Request) (*Session, error) {
	if req.Folder == nil || req.Job == nil {
		return nil, errors.New("a run needs a folder and a job")
	}
	if len(req.Args) > 0 && !req.Job.TakesArgs {
		return nil, ErrNoArgs
	}
	outOfHours, err := m.gate(req)
	if err != nil {
		return nil, err
	}
	run := req.newRun()
	if run.Parallel() && run.ID == "" {
		// The run's own lock takes its ID.
		run.ID = model.NewID()
	}
	if outOfHours {
		run.Reason = strings.TrimSpace(req.Reason)
	}
	if current, err := m.acquire(run.LockKey(), run.Folder, req.Force); err != nil {
		if req.Queue && req.ID == "" && errors.Is(err, ErrBusy) {
			return m.enqueue(ctx, req, run)
		}
		return current, err
	}
	if current, err := m.replace(ctx, run); err != nil {
		m.release(run.LockKey(), nil)
		return current, err
	}
	if err := m.opts.Runs.Create(ctx, run); err != nil {
		m.release(run.LockKey(), nil)
		return nil, err
	}
	session := m.newSession(ctx, req, run)
	m.mu.Lock()
	m.sessions[run.ID] = session
	m.byKey[run.LockKey()] = session
	m.mu.Unlock()
	m.publish(*run)
	go m.serve(session)
	return session, nil
}

// newSession is the session of a run a request starts, with what the
// run needs of the request.
func (m *Manager) newSession(ctx context.Context, req Request, run *model.Run) *Session {
	session := NewSession(run, req.Folder, req.Job)
	session.env = req.Env
	session.dir, session.args, session.timeout = req.Dir, req.Args, req.Timeout
	session.Titled = m.publish
	session.workspaceName, session.workspace = m.workspace(ctx, req)
	session.workspaceID = m.workspaceID(ctx, session.workspaceName)
	if req.User != nil {
		session.role = req.User.EffectiveRole()
		if m.opts.Users != nil {
			session.permissions = m.opts.Users.Permissions(ctx, req.User).Summary()
		}
	}
	return session
}

// RunsTopic is the topic the manager publishes every run on.
const RunsTopic = "runs"

// publish tells the runs topic how a run stands, when there is a bus.
func (m *Manager) publish(run model.Run) {
	if m.Events != nil {
		m.Events.Publish(RunsTopic, "run", run)
	}
}

// Stop ends a run as stopped; a queued run leaves the queue, stopped.
func (m *Manager) Stop(id string) error {
	session, ok := m.Get(id)
	if !ok {
		return ErrNotFound
	}
	if m.dequeue(session, Status{State: model.StateStopped}) {
		return nil
	}
	session.Stop(m.opts.Grace)
	return nil
}

// Kill ends a run as killed; a queued run leaves the queue, killed.
func (m *Manager) Kill(id string) error {
	session, ok := m.Get(id)
	if !ok {
		return ErrNotFound
	}
	if m.dequeue(session, Status{State: model.StateKilled}) {
		return nil
	}
	session.Kill(m.opts.Grace)
	return nil
}

// Shutdown ends the queued runs as failed, kills every run and waits
// for the sessions to settle.
func (m *Manager) Shutdown(ctx context.Context) {
	m.mu.Lock()
	m.closing = true
	var waiting []*Session
	for _, queue := range m.queued {
		waiting = append(waiting, queue...)
	}
	m.mu.Unlock()
	for _, session := range waiting {
		m.dequeue(session, Status{State: model.StateFailed, ExitCode: 1, Message: model.QueuedLost})
	}
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
	// that runs again on the status finds the job free, or to the run
	// that waits next.
	m.handOff(run.LockKey(), session)
	m.settle(ctx, session, status)
	go m.forget(session)
}

// settle writes the outcome down and tells the windows.
func (m *Manager) settle(ctx context.Context, session *Session, status Status) {
	// The session takes the outcome before the row does: a reader that
	// sees the row ended and then reads the session, as the API does while
	// the manager keeps the session, must not find it still running.
	session.mu.Lock()
	session.Run.State, session.Run.ExitCode = status.State, int64(status.ExitCode)
	session.mu.Unlock()
	run := session.Snapshot()
	if err := m.opts.Runs.Finish(ctx, &run); err != nil {
		slog.Warn("run outcome not recorded", "run", run.ID, "err", err)
	}
	session.mu.Lock()
	session.Run.FinishedAt = run.FinishedAt
	session.mu.Unlock()
	m.collectArtifacts(ctx, session)
	slog.Info("run finished", "folder", run.Folder, "job", run.Job, "user", run.UserName, "state", status.State, "exit", status.ExitCode)
	m.publish(run)
	session.finish(status)
}

// execute does the run: a pull when the folder is in a checkout, the
// keys the policies grant the run's user in an agent, the env files, then
// the process, pumped until it exits.
func (m *Manager) execute(ctx context.Context, session *Session) Status {
	f, job, run := session.Folder, session.Job, session.Run
	if len(job.Command) == 0 {
		return failed(session, "the job has no command")
	}
	if err := m.pull(ctx, session); err != nil {
		return failed(session, err.Error())
	}
	keys, err := m.agent(ctx, session)
	if err != nil {
		return failed(session, err.Error())
	}
	defer keys.Close()
	if m.opts.Env != nil {
		if err := m.opts.Env.Generate(ctx, f); err != nil {
			return failed(session, "generating env files: "+err.Error())
		}
	}
	cmd, err := m.command(ctx, session)
	if err != nil {
		return failed(session, err.Error())
	}
	// The run's agent or none: never the server's own SSH_AUTH_SOCK.
	socket := ""
	if keys != nil {
		socket = keys.Socket
	}
	cmd.Env = append(cmd.Env, "SSH_AUTH_SOCK="+socket)
	proc, err := StartProcess(cmd, int(run.TermCols), int(run.TermRows))
	if err != nil {
		return failed(session, "cannot start: "+err.Error())
	}
	session.setProcess(proc, m.opts.Grace)
	if session.timeout > 0 {
		limit := time.AfterFunc(session.timeout, func() {
			fmt.Fprintf(session, "\r\n\x1b[31mthe run went past %s\x1b[0m\r\n", session.timeout)
			session.Kill(m.opts.Grace)
		})
		defer limit.Stop()
	}
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
// job's argv passed as arguments so the shell never parses a name, and
// the request's args after --, the same way; the secrets and the run
// facts in the environment, the artifact directory ready. The process
// starts in the request's directory, or the folder's.
func (m *Manager) command(ctx context.Context, session *Session) (*exec.Cmd, error) {
	f, job, run := session.Folder, session.Job, session.Run
	args := append([]string{"-l", "-c", `exec "$0" "$@"`}, job.Command...)
	// A host job's command carries its args already, as the remote
	// command of ssh.
	if len(session.args) > 0 && !job.IsHost() {
		args = append(append(args, "--"), session.args...)
	}
	cmd := exec.Command("bash", args...)
	cmd.Dir = f.Dir
	if session.dir != "" {
		cmd.Dir = session.dir
	}
	cmd.Env = inherited(os.Environ())
	// Every run gets its folder's secrets, Claude's built-ins too, and
	// the log of each secret records the run.
	if m.opts.Secrets != nil {
		secrets, err := m.opts.Secrets.Reaching(ctx, f.ID)
		if err != nil {
			return nil, err
		}
		reads := make([]model.SecretRead, 0, len(secrets))
		for _, secret := range secrets {
			cmd.Env = append(cmd.Env, secret.Name+"="+secret.Content)
			reads = append(reads, model.SecretRead{SecretID: secret.ID, Name: secret.Name, RunID: run.ID, Folder: run.Folder, Dir: run.Dir, Job: run.Job, UserName: run.UserName})
		}
		if m.opts.SecretReads != nil {
			if err := m.opts.SecretReads.Record(ctx, reads); err != nil {
				return nil, err
			}
		}
	}
	if m.opts.Integrations != nil {
		vars, err := m.opts.Integrations.Env(ctx, run.UserName, session.role)
		if err != nil {
			return nil, err
		}
		for _, key := range model.SortedKeys(vars) {
			cmd.Env = append(cmd.Env, key+"="+vars[key])
		}
	}
	cmd.Env = append(cmd.Env, session.env...)
	cmd.Env = append(cmd.Env,
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"COLUMNS="+strconv.Itoa(int(run.TermCols)),
		"LINES="+strconv.Itoa(int(run.TermRows)),
		"TASK_UI=1",
		"TASK_UI_USER="+run.UserName,
		"TASK_UI_ROLE="+session.role,
		"TASK_UI_FOLDER="+f.ID,
		"TASK_UI_JOB="+job.ID,
		"TASK_UI_RUN_ID="+run.ID,
		"TASK_UI_HOST_PWD="+m.hostPath(cmd.Dir),
	)
	if m.opts.HostData != "" {
		cmd.Env = append(cmd.Env, "TASK_UI_HOST_DATA="+m.opts.HostData)
	}
	if session.permissions != "" {
		cmd.Env = append(cmd.Env, "TASK_UI_PERMISSIONS="+session.permissions)
	}
	if session.workspace != "" {
		cmd.Env = append(cmd.Env,
			"TASK_UI_WORKSPACE="+session.workspaceName,
			"TASK_UI_WORKSPACE_DIR="+session.workspace,
		)
	}
	if dir := m.artifactDir(run); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		cmd.Env = append(cmd.Env, "TASK_UI_ARTIFACTS="+dir)
	}
	return cmd, nil
}

// workspace returns the folder and the directory of the request's
// workspace: the one it names, or the workspace of the first segment
// of WorkspaceOf, else of the folder, under the root, or the default
// workspace, _ at the root, for a folder in no named one. Empty with no
// workspace known.
func (m *Manager) workspace(ctx context.Context, req Request) (string, string) {
	if req.Workspace != "" {
		if m.opts.Root != "" && filepath.Clean(req.Workspace) == filepath.Clean(m.opts.Root) {
			return model.Default, req.Workspace
		}
		return filepath.Base(req.Workspace), req.Workspace
	}
	of := req.Folder.ID
	if req.WorkspaceOf != "" {
		of = req.WorkspaceOf
	}
	if m.opts.Workspaces == nil {
		return "", ""
	}
	top, _, _ := strings.Cut(of, "/")
	if top != model.Default {
		if _, err := m.opts.Workspaces.GetByFolder(ctx, top); err == nil {
			return top, filepath.Join(m.opts.Root, top)
		}
	}
	if _, err := m.opts.Workspaces.GetByFolder(ctx, model.Default); err != nil {
		return "", ""
	}
	return model.Default, m.opts.Root
}

// WorkspaceID returns the ID of the workspace row a run of the request
// takes, as Start finds it: "" for an internal folder's own run, or
// with no workspace known.
func (m *Manager) WorkspaceID(ctx context.Context, req Request) string {
	name, _ := m.workspace(ctx, req)
	return m.workspaceID(ctx, name)
}

// workspaceID returns the ID of the workspace row of a folder, "" for
// none or with no workspaces known.
func (m *Manager) workspaceID(ctx context.Context, name string) string {
	if name == "" || m.opts.Workspaces == nil {
		return ""
	}
	workspace, err := m.opts.Workspaces.GetByFolder(ctx, name)
	if err != nil {
		return ""
	}
	return workspace.ID
}

func (m *Manager) artifactDir(run *model.Run) string {
	if m.opts.ArtifactRoot == "" {
		return ""
	}
	return filepath.Join(m.opts.ArtifactRoot, run.ID)
}
