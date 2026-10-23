package api

import (
	"errors"
	"net/http"
	"net/url"

	"golang.org/x/net/websocket"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/runner"
)

// The attach socket speaks two kinds of frames. Binary frames carry the
// terminal output. Text frames carry JSON control messages:
//
//	server: {"type":"status","state":"running|ok|failed|stopped|killed","exit_code":0,"message":""}
//	server: {"type":"input","state":"on|off","message":""}
//	client: {"type":"input","message":"..."} | {"type":"resize","cols":120,"rows":30} | {"type":"stop"} | {"type":"kill"}
//	client: {"type":"take"} | {"type":"release"}
//
// Take turns a window on a running run of a job that takes no input
// into one that types into it, for a user with the input permission;
// release turns it back. The server answers either with an input
// message, its state on or off and why when refused.
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

// attach serves the websocket a window attaches to a run with. The window
// may send input when the user may run the job; everyone who may see the
// run may watch it.
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

// serveAttach finds the run the socket names and, while its session is
// there, opens a window on it. A run that is over gets its last status
// and nothing more.
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
	f, job, _, _ := h.targetIn(r, run.Folder, run.Dir, run.Job)
	w := &Window{Terminal: session, Interactive: session.Job != nil && session.Job.Interactive, ReadOnly: !h.canRun(r, f, job)}
	if user := access.UserFrom(r.Context()); user != nil && h.takesInput(r, session) {
		w.Taker = user.Name
	}
	h.window(ws, session, w)
}

// takesInput reports whether the user of a request may take input on
// a session: a job that takes none, and the input permission.
func (h *Handlers) takesInput(r *http.Request, session *runner.Session) bool {
	return session.Job != nil && !session.Job.Interactive && h.can(r, "input")
}

// window pumps a session to the socket and the socket's control messages
// to the session, until the run ends or the window closes.
func (h *Handlers) window(ws *websocket.Conn, session *runner.Session, w *Window) {
	sub := session.Attach(w.ReadOnly)
	defer session.Detach(sub)

	// The reads end when the window leaves; a quiet run would not
	// notice otherwise.
	gone := make(chan struct{})
	go func() {
		defer close(gone)
		readControls(ws, w)
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

// readControls applies what the window sends, and answers what it
// answers, until the window goes.
func readControls(ws *websocket.Conn, w *Window) {
	for {
		var frame []byte
		if err := websocket.Message.Receive(ws, &frame); err != nil {
			return
		}
		if reply := w.Apply(frame); reply != nil {
			_ = websocket.Message.Send(ws, string(reply))
		}
	}
}

func sendControl(ws *websocket.Conn, msg control) {
	_ = websocket.Message.Send(ws, string(controlFrame(msg)))
}
