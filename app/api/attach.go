package api

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/websocket"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// The attach socket speaks two kinds of frames. Binary frames carry the
// terminal output. Text frames carry JSON control messages:
//
//	server: {"type":"status","state":"running|ok|failed|stopped|killed","exit_code":0,"message":""}
//	server: {"type":"input","state":"on|off","message":""}
//	server: {"type":"title","title":"..."}
//	server: {"type":"chain","chain":{"id":"...","kind":"summary","position":1,"total":2,"next":"...","complete":false}}
//	client: {"type":"input","message":"..."} | {"type":"resize","cols":120,"rows":30} | {"type":"stop"} | {"type":"kill"}
//	client: {"type":"take"} | {"type":"release"}
//
// Take turns a window on a running run of a job that takes no input
// into one that types into it, for a user with the input permission;
// release turns it back. The server answers either with an input
// message, its state on or off and why when refused.
//
// A title message names what the run's terminal set as its title, on
// attach and on every change; no title clears it.
//
// A chain message tells a window on a run of a chain, a Run and
// summarize, a workflow run or a batch of host runs, where the run
// stands in it: on attach, and after the status of its end once the
// chain has the next run or is complete. The socket stays open after
// the end for that, while the chain waits for the next run.
//
// Closing the socket detaches the window; the run goes on.
type control struct {
	Type     string    `json:"type"`
	State    string    `json:"state,omitempty"`
	ExitCode int       `json:"exit_code,omitempty"`
	Message  string    `json:"message,omitempty"`
	Title    string    `json:"title,omitempty"`
	Cols     int       `json:"cols,omitempty"`
	Rows     int       `json:"rows,omitempty"`
	Chain    *RunChain `json:"chain,omitempty"`
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
	ctx := r.Context()
	session, ok := h.opts.Runner.Get(run.ID)
	if !ok {
		sendControl(ws, control{Type: "status", State: run.State, ExitCode: int(run.ExitCode), Message: "the run is over, see its recording"})
		h.followChain(ws, *run, watchGone(ws, nil))
		return
	}
	if c := h.chain(ctx, *run); c != nil {
		sendControl(ws, control{Type: "chain", Chain: c})
	}
	f, job, _, _ := h.targetIn(r, run.Folder, run.Dir, run.Job)
	w := &Window{Terminal: session, Interactive: session.Job != nil && session.Job.Interactive, ReadOnly: !h.canRun(r, f, job)}
	if user := access.UserFrom(r.Context()); user != nil && h.takesInput(r, session) {
		w.Taker = user.Name
	}
	gone := watchGone(ws, w)
	if h.window(ws, session, w, gone) {
		h.followChain(ws, session.Snapshot(), gone)
	}
}

// watchGone reads the window's control messages until the window goes,
// which the channel tells: applied to w, or dropped for a run that is
// over, nil.
func watchGone(ws *websocket.Conn, w *Window) <-chan struct{} {
	gone := make(chan struct{})
	go func() {
		defer close(gone)
		if w != nil {
			readControls(ws, w)
			return
		}
		var frame []byte
		for websocket.Message.Receive(ws, &frame) == nil {
		}
	}()
	return gone
}

// followChain sends the window the chain of a run that ended, once the
// chain has the next run or is complete, as FollowChain waits for it.
func (h *Handlers) followChain(ws *websocket.Conn, run model.Run, gone <-chan struct{}) {
	ctx := ws.Request().Context()
	tick := time.NewTicker(chainPoll)
	defer tick.Stop()
	timeout := time.NewTimer(chainWait)
	defer timeout.Stop()
	FollowChain(func() *RunChain { return h.chain(ctx, run) }, func(c *RunChain) error {
		return websocket.Message.Send(ws, string(controlFrame(control{Type: "chain", Chain: c})))
	}, tick.C, gone, timeout.C)
}

// takesInput reports whether the user of a request may take input on
// a session: a job that takes none, and the input permission.
func (h *Handlers) takesInput(r *http.Request, session *runner.Session) bool {
	return session.Job != nil && !session.Job.Interactive && h.can(r, "input")
}

// window pumps a session to the socket, until the run ends or the
// window closes, which gone tells: the reads of the socket's control
// messages end when the window leaves, and a quiet run would not
// notice otherwise. It reports whether the run ended.
func (h *Handlers) window(ws *websocket.Conn, session *runner.Session, w *Window, gone <-chan struct{}) bool {
	sub := session.Attach(w.ReadOnly)
	defer session.Detach(sub)

	for {
		select {
		case <-gone:
			return false
		case event, ok := <-sub.C:
			if !ok {
				return false
			}
			frame, text := w.Frame(event)
			var err error
			if text {
				err = websocket.Message.Send(ws, string(frame))
			} else {
				err = websocket.Message.Send(ws, frame)
			}
			if err != nil {
				return false
			}
			if event.Status != nil {
				return true
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
