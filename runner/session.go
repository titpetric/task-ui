package runner

import (
	"io"
	"sync"
	"time"

	"github.com/hinshun/vt10x"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// replaySize is how much of the output a late window gets.
const replaySize = 256 << 10

// TitleEvery is how often at most a session tells Titled of a title
// that keeps changing: the last title always arrives.
const TitleEvery = 250 * time.Millisecond

// pendingSize bounds the input kept for a process that has not started.
const pendingSize = 64 << 10

// Session is one run in progress, and what remains of it after: the
// process on its pty, the terminal emulator that follows the output, the
// windows attached, and how it ended. A session outlives the window that
// started it: windows attach and detach, the run goes on. Windows attach
// before the process starts and input waits for it; the manager starts the
// process, and a session it never starts stays running. Every run has its
// pty on stdin, so a job that takes no input still answers a prompt once a
// user takes its input.
type Session struct {
	Run    *model.Run
	Folder *folder.Folder
	Job    *folder.Job
	// role is the role of the user who started the run, for the instance
	// keys its runs get.
	role string
	// permissions say in a line what that user may do.
	permissions string
	// env is what the request adds to the run's environment.
	env []string
	// dir is where the process starts, "" for the folder's directory,
	// and args what follows the job's argv after --.
	dir  string
	args []string
	// timeout kills the run once it runs that long; zero lets it run.
	timeout time.Duration
	// workspace is the directory of the run's workspace, "" for none, and
	// workspaceName its folder: internal for the default workspace, the root.
	// workspaceID is its row's ID, whose keys the run gets; "" without
	// a row.
	workspace     string
	workspaceName string
	workspaceID   string
	// noWorkspace says the run has no workspace: no workspace secrets or
	// keys, and no pull of a workspace.
	noWorkspace bool
	// home is the directory of the workspace's home project, the run's
	// HOME; "" keeps the server's.
	home string
	// Titled hears the run row when the terminal's title changes, at
	// most every TitleEvery with the title of the moment; nil tells
	// nobody. The manager publishes it on the runs topic.
	Titled func(run model.Run)

	mu   sync.Mutex
	term vt10x.Terminal
	tail *ring
	proc *Process
	// pending is the input typed before the process started, and the
	// size the window requested; both reach the process when it starts.
	pending    []byte
	cols, rows int
	// inputBy names who took input on a job that takes none; "" until
	// someone does.
	inputBy string
	// titles follows the output for the title the terminal sets, which
	// Run.Title keeps; titleAt is when Titled last heard of it, and
	// titleDue says a later call carries the changes since.
	titles   titleScanner
	titleAt  time.Time
	titleDue bool
	subs     map[*Subscriber]struct{}
	sinks    []io.Writer
	status   Status
	// slot says the run takes one of the instance's slots while it runs.
	slot    bool
	stopped bool
	killed  bool
	done    chan struct{}
}

// NewSession returns the session of a run whose process has not started, for windows to attach to.
func NewSession(run *model.Run, f *folder.Folder, job *folder.Job) *Session {
	return &Session{
		Run:    run,
		Folder: f,
		Job:    job,
		term:   newTerminal(int(run.TermCols), int(run.TermRows)),
		tail:   newRing(replaySize),
		subs:   map[*Subscriber]struct{}{},
		status: Status{State: model.StateRunning},
		done:   make(chan struct{}),
	}
}

// ID is the run id.
func (s *Session) ID() string {
	return s.Run.ID
}

// Key is the lock key: one session per job per folder and directory, or per run for a parallel job.
func (s *Session) Key() string {
	return s.Run.LockKey()
}

// Done returns a channel that closes once the run ended and the session settled.
func (s *Session) Done() <-chan struct{} {
	return s.done
}

// Status returns how the run is doing.
func (s *Session) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

// Frame returns what the terminal shows now.
func (s *Session) Frame() Frame {
	return snapshot(s.term)
}

// Attach adds a window that gets what the run showed so far, then the output as it happens, then the final status.
// A window on a finished run gets the replay and the status at once.
func (s *Session) Attach(readOnly bool) *Subscriber {
	sub := newSubscriber(64, readOnly)
	s.mu.Lock()
	defer s.mu.Unlock()
	if replay := s.tail.Bytes(); len(replay) > 0 {
		sub.send(Event{Output: replay})
	}
	if title := s.Run.Title; title != "" {
		sub.send(Event{Title: &title})
	}
	if s.status.Finished() {
		status := s.status
		sub.send(Event{Status: &status})
		sub.close()
		return sub
	}
	s.subs[sub] = struct{}{}
	return sub
}

// Detach removes a window. The run goes on.
func (s *Session) Detach(sub *Subscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.subs[sub]; ok {
		delete(s.subs, sub)
		sub.close()
	}
}

// Attached returns the number of attached windows.
func (s *Session) Attached() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.subs)
}

// Input sends keystrokes to the process, for an interactive job or one
// whose input someone took.
func (s *Session) Input(p []byte) error {
	s.mu.Lock()
	if !s.Job.Interactive && s.inputBy == "" {
		s.mu.Unlock()
		return ErrNotInteractive
	}
	proc := s.proc
	if proc == nil {
		if s.status.Finished() {
			s.mu.Unlock()
			return ErrNotRunning
		}
		// Typed before the process is up: keep it for the process, a
		// window cannot tell the two moments apart.
		if len(s.pending)+len(p) <= pendingSize {
			s.pending = append(s.pending, p...)
		}
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()
	_, err := proc.Write(p)
	return err
}

// Resize changes the terminal size for the process and the emulator.
func (s *Session) Resize(cols, rows int) {
	if cols <= 0 || rows <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.term.Resize(cols, rows)
	s.cols, s.rows = cols, rows
	if s.proc != nil {
		s.proc.Resize(cols, rows)
	}
}

// Stop ends the run as stopped: what the Stop button does.
func (s *Session) Stop(grace time.Duration) {
	s.end(true, grace)
}

// Kill ends the run as killed: what Run force and Kill do.
func (s *Session) Kill(grace time.Duration) {
	s.end(false, grace)
}

// end signals the process to stop. A session whose process has not started
// yet remembers, and stops it as soon as it starts.
func (s *Session) end(byUser bool, grace time.Duration) {
	s.mu.Lock()
	if byUser {
		s.stopped = true
	} else {
		s.killed = true
	}
	proc := s.proc
	s.mu.Unlock()
	if proc != nil {
		proc.Stop(grace)
	}
}

// addSink adds a writer that receives everything the run shows.
func (s *Session) addSink(w io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sinks = append(s.sinks, w)
}

// Title returns the title the run's terminal set, "" for none.
func (s *Session) Title() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Run.Title
}

// Snapshot returns a copy of the run row as the session knows it.
func (s *Session) Snapshot() model.Run {
	s.mu.Lock()
	defer s.mu.Unlock()
	return *s.Run
}

// write broadcasts output: to the emulator, the replay tail, the sinks
// (the recording), and every attached window.
func (s *Session) write(p []byte) {
	s.mu.Lock()
	_, _ = s.term.Write(p)
	_, _ = s.tail.Write(p)
	for _, sink := range s.sinks {
		_, _ = sink.Write(p)
	}
	event := Event{Output: append([]byte{}, p...)}
	for sub := range s.subs {
		sub.send(event)
	}
	var titled *model.Run
	if title, changed := s.titles.scan(p); changed {
		s.Run.Title = title
		for sub := range s.subs {
			sub.send(Event{Title: &title})
		}
		titled = s.retitled()
	}
	s.mu.Unlock()
	if titled != nil {
		s.Titled(*titled)
	}
}

// retitled decides when Titled hears of the title: it returns the run
// to tell it of now, when it heard nothing for TitleEvery, else nil and
// tells it once that passed, with the title of then. The caller holds
// the lock.
func (s *Session) retitled() *model.Run {
	if s.Titled == nil || s.titleDue {
		return nil
	}
	wait := TitleEvery - time.Since(s.titleAt)
	if wait <= 0 {
		s.titleAt = time.Now()
		run := *s.Run
		return &run
	}
	s.titleDue = true
	time.AfterFunc(wait, func() {
		s.mu.Lock()
		s.titleDue, s.titleAt = false, time.Now()
		run := *s.Run
		s.mu.Unlock()
		s.Titled(run)
	})
	return nil
}

// Write implements io.Writer: the session is the run's terminal.
func (s *Session) Write(p []byte) (int, error) {
	s.write(p)
	return len(p), nil
}

// setProcess installs the running process, stopping it at once when the
// session ended before it started.
func (s *Session) setProcess(proc *Process, grace time.Duration) {
	s.mu.Lock()
	s.proc = proc
	ended := s.stopped || s.killed
	pending, cols, rows := s.pending, s.cols, s.rows
	s.pending = nil
	s.mu.Unlock()
	if ended {
		proc.Stop(grace)
		return
	}
	if cols > 0 && rows > 0 {
		proc.Resize(cols, rows)
	}
	if len(pending) > 0 {
		_, _ = proc.Write(pending)
	}
}

// finish records the status, tells every window, and settles.
func (s *Session) finish(status Status) {
	s.mu.Lock()
	s.status = status
	s.Run.State = status.State
	s.Run.ExitCode = int64(status.ExitCode)
	s.proc = nil
	subs := s.subs
	s.subs = map[*Subscriber]struct{}{}
	s.mu.Unlock()
	for sub := range subs {
		sub.send(Event{Status: &status})
		sub.close()
	}
	close(s.done)
}

// outcome decides the final state from how the process ended and who
// requested it.
func (s *Session) outcome(exitCode int) Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch {
	case s.killed:
		return Status{State: model.StateKilled, ExitCode: exitCode}
	case s.stopped:
		return Status{State: model.StateStopped, ExitCode: exitCode}
	case exitCode != 0:
		return Status{State: model.StateFailed, ExitCode: exitCode}
	}
	return Status{State: model.StateOK}
}

// Default colours, which a theme maps to its own ink and paper.
const (
	// DefaultFG is the terminal's default foreground.
	DefaultFG = uint32(vt10x.DefaultFG)
	// DefaultBG is the terminal's default background.
	DefaultBG = uint32(vt10x.DefaultBG)
)

// The vt10x attribute bits a frame reads, in the order vt10x declares
// them.
const (
	attrReverse = 1 << iota
	attrUnderline
	attrBold
)

// snapshot reads a frame off a terminal.
// newTerminal returns a terminal emulator of a size: vt10x stays imported
// in this file alone.
func newTerminal(cols, rows int) vt10x.Terminal {
	return vt10x.New(vt10x.WithSize(cols, rows))
}

func snapshot(term vt10x.Terminal) Frame {
	term.Lock()
	defer term.Unlock()
	cols, rows := term.Size()
	frame := Frame{Cols: cols, Rows: rows, Title: term.Title()}
	cursor := term.Cursor()
	frame.CursorX, frame.CursorY = cursor.X, cursor.Y
	for y := 0; y < rows; y++ {
		line := make([]Cell, cols)
		for x := 0; x < cols; x++ {
			glyph := term.Cell(x, y)
			line[x] = Cell{
				Char:    glyph.Char,
				FG:      uint32(glyph.FG),
				BG:      uint32(glyph.BG),
				Bold:    glyph.Mode&attrBold != 0,
				Reverse: glyph.Mode&attrReverse != 0,
			}
		}
		frame.Lines = append(frame.Lines, line)
	}
	return frame
}
