package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/net/websocket"

	"github.com/titpetric/task-ui/server/internal"
	"github.com/titpetric/task-ui/server/model"
	"github.com/titpetric/task-ui/server/repository"
)

// The run socket speaks two kinds of frames. Binary frames carry terminal
// output. Text frames carry JSON control messages:
//
//	server: {"type":"status","state":"running|busy|denied|error|ok|failed|stopped|killed", ...}
//	client: {"type":"resize","cols":120,"rows":30} | {"type":"stop"} | {"type":"input","message":"..."}
//
// Input is a control frame rather than raw bytes, so a key the user types
// is never mistaken for a message and the server never has to guess.
//
// Closing the socket stops the run: a run belongs to the window that started it.
type control struct {
	Type    string `json:"type"`
	State   string `json:"state,omitempty"`
	Message string `json:"message,omitempty"`
	ID      string `json:"id,omitempty"`
	User    string `json:"user,omitempty"`
	Started string `json:"started,omitempty"`
	Exit    *int   `json:"exit,omitempty"`
	Cols    int    `json:"cols,omitempty"`
	Rows    int    `json:"rows,omitempty"`
}

// socket serializes writes to a websocket connection.
type socket struct {
	mu sync.Mutex
	ws *websocket.Conn
}

func (s *socket) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := websocket.Message.Send(s.ws, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *socket) status(msg control) {
	msg.Type = "status"
	data, _ := json.Marshal(msg)
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = websocket.Message.Send(s.ws, string(data))
}

// RunSocket runs a job over a websocket.
func (svc *Server) RunSocket() http.Handler {
	return websocket.Server{
		Handshake: func(config *websocket.Config, r *http.Request) error {
			// Refuse cross site sockets: the cookie would let another
			// site start jobs in the user's name.
			origin, err := url.Parse(r.Header.Get("Origin"))
			if err != nil || !sameHost(origin.Host, r) {
				return errors.New("cross origin websocket refused")
			}
			config.Origin = origin
			return nil
		},
		Handler: svc.runSocket,
	}
}

func (svc *Server) runSocket(ws *websocket.Conn) {
	defer ws.Close()
	ws.PayloadType = websocket.BinaryFrame

	r := ws.Request()
	out := &socket{ws: ws}
	user := currentUser(r)

	data := svc.pageData(r, "")
	project := svc.visibleProject(r, data)
	job := project.Job(chi.URLParam(r, "job"))
	if job == nil {
		out.status(control{State: "error", Message: "No such job."})
		return
	}
	if !user.CanRun(project, job) {
		out.status(control{State: "denied", Message: "Your role may view this job, but not run it."})
		return
	}

	query := r.URL.Query()
	force := query.Get("force") == "1"
	cols, _ := strconv.Atoi(query.Get("cols"))
	rows, _ := strconv.Atoi(query.Get("rows"))
	if cols < 20 || cols > 500 {
		cols = svc.config.Columns
	}
	if rows < 5 || rows > 200 {
		rows = svc.config.Rows
	}

	started := time.Now()
	run := &internal.Run{
		ID:      repository.NewID(started),
		Key:     job.Key(),
		Project: project.ID,
		Job:     job.Name,
		User:    user.Name,
		Started: started,
	}

	if force {
		if current, ok := svc.runs.Get(run.Key); ok {
			out.Write([]byte(fmt.Sprintf("\x1b[33mStopping the run %s started at %s...\x1b[0m\r\n", current.User, current.Started.Local().Format("15:04:05"))))
		}
	}
	if current, err := svc.runs.Acquire(run, force); err != nil {
		out.status(control{
			State:   "busy",
			Message: fmt.Sprintf("Already running, started by %s %s. Run force stops that run and starts again.", current.User, ago(current.Started)),
			User:    current.User,
			Started: current.Started.UTC().Format(time.RFC3339),
		})
		return
	}
	defer svc.runs.Release(run)

	slog.Info("run started", "project", project.ID, "job", job.Name, "user", user.Name, "force", force)

	state, exitCode, err := svc.execute(out, ws, project, job, run, cols, rows)
	if err != nil {
		slog.Error("run failed to start", "project", project.ID, "job", job.Name, "err", err)
		out.Write([]byte("\x1b[31m" + err.Error() + "\x1b[0m\r\n"))
		out.status(control{State: "error", Message: err.Error()})
		return
	}
	slog.Info("run finished", "project", project.ID, "job", job.Name, "user", user.Name, "state", state, "exit", exitCode, "duration", time.Since(started).Round(time.Millisecond))
	out.status(control{State: state, Exit: &exitCode, ID: run.ID})
}

// execute runs the job command on a pty and pumps it to the socket and
// the recording. It returns how the run ended.
func (svc *Server) execute(out *socket, ws *websocket.Conn, project *model.Project, job *model.Job, run *internal.Run, cols, rows int) (string, int, error) {
	if len(project.Command) == 0 {
		return "", 0, errors.New("project has no command configured")
	}

	// bash -l gives the login environment (PATH, profile), and passing the
	// command as arguments keeps the job name out of shell parsing.
	args := append([]string{"-l", "-c", `exec "$0" "$@"`}, project.Command...)
	args = append(args, job.Name)
	cmd := exec.Command("bash", args...)
	cmd.Dir = project.Dir

	// The env files the pipeline reads are generated with the secrets
	// first, and the secrets go into the process environment as well, so
	// a pipeline without env files gets them too.
	if err := svc.env.Generate(project); err != nil {
		return "", 0, fmt.Errorf("generating env files: %w", err)
	}
	secrets := svc.env.SecretsFor(project.ID)
	cmd.Env = os.Environ()
	for key, value := range secrets {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	cmd.Env = append(cmd.Env,
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"COLUMNS="+strconv.Itoa(cols),
		"LINES="+strconv.Itoa(rows),
		"TASK_UI=1",
		"TASK_UI_USER="+run.User,
		"TASK_UI_FOLDER="+project.ID,
		"TASK_UI_PROJECT="+project.ID,
		"TASK_UI_RUNBOOK="+job.Name,
		"TASK_UI_RUN_ID="+run.ID,
	)

	proc, err := internal.Start(cmd, cols, rows)
	if err != nil {
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
	run.SetStop(func() { stop(false) })

	out.status(control{State: model.StateRunning, ID: run.ID, User: run.User, Started: run.Started.UTC().Format(time.RFC3339)})

	var recorder *repository.Recorder
	if svc.config.History {
		recorder, err = svc.history.Start(model.HistoryRecord{
			ID:      run.ID,
			Project: project.ID,
			Job:     job.Name,
			User:    run.User,
			Started: run.Started,
			State:   model.StateRunning,
			Forced:  run.Forced,
		}, cols, rows)
		if err != nil {
			slog.Warn("history disabled for run", "err", err)
			recorder = nil
		}
	}

	// Client to process: input for interactive jobs, resize and stop
	// for all. The read fails when the window closes, which stops the run.
	go func() {
		for {
			var frame []byte
			if err := websocket.Message.Receive(ws, &frame); err != nil {
				stop(true)
				return
			}
			var msg control
			if json.Unmarshal(frame, &msg) != nil {
				continue
			}
			switch msg.Type {
			case "resize":
				proc.Resize(msg.Cols, msg.Rows)
			case "stop":
				out.Write([]byte("\r\n\x1b[33mStopping...\x1b[0m\r\n"))
				stop(true)
			case "input":
				if job.Interactive {
					proc.TTY.Write([]byte(msg.Message))
				}
			}
		}
	}()

	// Process to client and recording.
	buf := make([]byte, 32*1024)
	for {
		n, err := proc.TTY.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			if recorder != nil {
				recorder.Write(chunk)
			}
			_, _ = out.Write(chunk)
		}
		if err != nil {
			break
		}
	}

	exitCode := proc.Wait()

	stopMu.Lock()
	byUser := stopped
	stopMu.Unlock()

	state := model.StateOK
	switch {
	case run.Killed():
		state = model.StateKilled
	case byUser:
		state = model.StateStopped
	case exitCode != 0:
		state = model.StateFailed
	}
	if recorder != nil {
		if err := recorder.Finish(state, exitCode); err != nil {
			slog.Warn("history write failed", "err", err)
		}
	}
	return state, exitCode, nil
}

// APIRuns lists the active runs the user can see.
func (svc *Server) APIRuns(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	runs := []*internal.Run{}
	for _, run := range svc.runs.List() {
		if user.CanSee(run.Project) {
			runs = append(runs, run)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(runs)
}

// APIStop kills an active run. Anyone allowed to run the job may stop
// it, which is the same permission Run force needs.
func (svc *Server) APIStop(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "")
	project := svc.visibleProject(r, data)
	job := project.Job(chi.URLParam(r, "job"))
	if job == nil {
		http.NotFound(w, r)
		return
	}
	if !data.User.CanRun(project, job) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if run, ok := svc.runs.Get(job.Key()); ok {
		slog.Info("run killed", "project", project.ID, "job", job.Name, "by", data.User.Name, "owner", run.User)
		run.Kill()
	}
	if r.Header.Get("Accept") == "application/json" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, safeBack(r.FormValue("back")), http.StatusSeeOther)
}
