package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	"github.com/titpetric/task-ui/history"
	"github.com/titpetric/task-ui/run"
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

// socket is the one place the websocket library is reached: it serializes
// writes, reads control frames, and hands the request to the run.
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

// receive reads the next control frame. It fails when the window closes.
func (s *socket) receive() (control, error) {
	for {
		var frame []byte
		if err := websocket.Message.Receive(s.ws, &frame); err != nil {
			return control{}, err
		}
		var msg control
		if json.Unmarshal(frame, &msg) == nil && msg.Type != "" {
			return msg, nil
		}
	}
}

func (s *socket) request() *http.Request {
	return s.ws.Request()
}

// runSocket runs a job over a websocket.
func (svc *Server) runSocket() http.Handler {
	return websocket.Server{
		Handshake: func(config *websocket.Config, r *http.Request) error {
			// Refuse cross site sockets: the cookie would let another
			// site start jobs in the user's name.
			origin, err := url.Parse(r.Header.Get("Origin"))
			if err != nil || !sameHost(r, origin.Host) {
				return errors.New("cross origin websocket refused")
			}
			config.Origin = origin
			return nil
		},
		Handler: func(ws *websocket.Conn) {
			defer ws.Close()
			ws.PayloadType = websocket.BinaryFrame
			svc.serveRun(&socket{ws: ws})
		},
	}
}

// serveRun resolves the job, takes the run lock and runs the job, telling
// the window what happened at every step.
func (svc *Server) serveRun(conn *socket) {
	r := conn.request()
	user := currentUser(r)

	data := svc.pageData(r, "")
	f := svc.visibleFolder(r, data)
	job := f.Job(param(r, "job"))
	if job == nil {
		conn.status(control{State: "error", Message: "No such job."})
		return
	}
	if !user.CanRun(f, job) {
		conn.status(control{State: "denied", Message: "Your role may view this job, but not run it."})
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
	active := &run.Run{
		ID:      history.NewID(started),
		Key:     job.Key(),
		Folder:  f.ID,
		Job:     job.ID,
		User:    user.Name,
		Started: started,
	}

	if force {
		if current, ok := svc.runs.Get(active.Key); ok {
			fmt.Fprintf(conn, "\x1b[33mStopping the run %s started at %s...\x1b[0m\r\n", current.User, current.Started.Local().Format("15:04:05"))
		}
	}
	if current, err := svc.runs.Acquire(active, force); err != nil {
		conn.status(control{
			State:   "busy",
			Message: fmt.Sprintf("Already running, started by %s %s. Run force stops that run and starts again.", current.User, ago(current.Started)),
			User:    current.User,
			Started: current.Started.UTC().Format(time.RFC3339),
		})
		return
	}
	defer svc.runs.Release(active)

	slog.Info("run started", "folder", f.ID, "job", job.ID, "user", user.Name, "force", force)

	state, exitCode, err := svc.execute(conn, f, job, active, cols, rows)
	if err != nil {
		slog.Error("run failed to start", "folder", f.ID, "job", job.ID, "err", err)
		fmt.Fprintf(conn, "\x1b[31m%s\x1b[0m\r\n", err)
		conn.status(control{State: "error", Message: err.Error()})
		return
	}
	slog.Info("run finished", "folder", f.ID, "job", job.ID, "user", user.Name, "state", state, "exit", exitCode, "duration", time.Since(started).Round(time.Millisecond))
	conn.status(control{State: state, Exit: &exitCode, ID: active.ID})
}
