package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/websocket"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/runner"
)

// The attach socket speaks two kinds of frames. Binary frames carry the
// terminal output. Text frames carry JSON control messages:
//
//	server: {"type":"status","state":"running|ok|failed|stopped|killed","exit_code":0,"message":""}
//	client: {"type":"input","message":"..."} | {"type":"resize","cols":120,"rows":30} | {"type":"stop"} | {"type":"kill"}
//
// Closing the socket detaches the window; the run goes on.
type control struct {
	Type     string `json:"type"`
	State    string `json:"state,omitempty"`
	ExitCode int    `json:"exit_code,omitempty"`
	Message  string `json:"message,omitempty"`
	Cols     int    `json:"cols,omitempty"`
	Rows     int    `json:"rows,omitempty"`
}

// attach serves the websocket a window attaches to a run with. Input is
// allowed when the user may run the job; everyone who may see the run
// may watch it.
func (h *Handlers) attach() http.Handler {
	return websocket.Server{
		Handshake: func(config *websocket.Config, r *http.Request) error {
			// Refuse cross site sockets: the cookie would let another
			// site drive jobs in the user's name.
			origin, err := url.Parse(r.Header.Get("Origin"))
			if err != nil || !access.SameHost(r, origin.Host) {
				return errors.New("cross origin websocket refused")
			}
			config.Origin = origin
			return nil
		},
		Handler: func(ws *websocket.Conn) {
			defer ws.Close()
			ws.PayloadType = websocket.BinaryFrame
			h.serveAttach(ws)
		},
	}
}

// serveAttach pumps a session to the socket and the socket's control
// messages to the session, until the run ends or the window closes.
func (h *Handlers) serveAttach(ws *websocket.Conn) {
	r := ws.Request()
	run, err := h.runOf(r)
	if err != nil {
		sendControl(ws, control{Type: "status", State: "error", Message: err.Error()})
		return
	}
	session, ok := h.opts.Runner.Get(run.ID)
	if !ok {
		sendControl(ws, control{Type: "status", State: run.State, ExitCode: int(run.ExitCode), Message: "the run is over, see its recording"})
		return
	}
	f, job, _ := h.target(r, run.Folder, run.Job)
	readOnly := !canRun(access.UserFrom(r.Context()), f, job)

	sub := session.Attach(readOnly)
	defer session.Detach(sub)

	// The window is gone when its reads end; a quiet run would not
	// notice otherwise.
	gone := make(chan struct{})
	go func() {
		defer close(gone)
		h.readControls(ws, session, readOnly)
	}()

	for {
		select {
		case <-gone:
			return
		case event, ok := <-sub.C:
			if !ok {
				return
			}
			if event.Status != nil {
				sendControl(ws, control{Type: "status", State: event.Status.State, ExitCode: event.Status.ExitCode, Message: event.Status.Message})
				return
			}
			if err := websocket.Message.Send(ws, event.Output); err != nil {
				return
			}
		}
	}
}

// readControls applies what the window sends: input, resizes, stop and
// kill. A read-only window's input and ends are ignored.
func (h *Handlers) readControls(ws *websocket.Conn, session *runner.Session, readOnly bool) {
	for {
		var frame []byte
		if err := websocket.Message.Receive(ws, &frame); err != nil {
			return
		}
		var msg control
		if json.Unmarshal(frame, &msg) != nil {
			continue
		}
		switch msg.Type {
		case "resize":
			session.Resize(msg.Cols, msg.Rows)
		case "input":
			if !readOnly {
				_ = session.Input([]byte(msg.Message))
			}
		case "stop":
			if !readOnly {
				session.Stop(5 * time.Second)
			}
		case "kill":
			if !readOnly {
				session.Kill(5 * time.Second)
			}
		}
	}
}

func sendControl(ws *websocket.Conn, msg control) {
	data, _ := json.Marshal(msg)
	_ = websocket.Message.Send(ws, string(data))
}
