//go:build integration

package api_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/websocket"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// message is a frame the attach socket sent: output, or a control.
type message struct {
	Type     string `json:"type"`
	State    string `json:"state"`
	ExitCode int    `json:"exit_code"`
	Message  string `json:"message"`
	Output   string `json:"-"`
}

// dial attaches a user to a run from an origin.
func (f *fixture) dial(t *testing.T, user, id, origin string) (*websocket.Conn, error) {
	t.Helper()
	url := "ws" + strings.TrimPrefix(f.server.URL, "http") + "/runs/" + id + "/attach"
	config, err := websocket.NewConfig(url, origin)
	assert.NoError(t, err)
	config.Header = http.Header{"X-User": {user}}
	return websocket.DialConfig(config)
}

// attach attaches a user to a run as the browser does.
func (f *fixture) attach(t *testing.T, user, id string) *websocket.Conn {
	t.Helper()
	ws, err := f.dial(t, user, id, f.server.URL)
	assert.NoError(t, err)
	t.Cleanup(func() { ws.Close() })
	return ws
}

// read returns the next frame of the socket.
func read(t *testing.T, ws *websocket.Conn) message {
	t.Helper()
	assert.NoError(t, ws.SetReadDeadline(time.Now().Add(10*time.Second)))
	var data []byte
	if err := websocket.Message.Receive(ws, &data); err != nil {
		t.Fatalf("reading the socket: %v", err)
	}
	var msg message
	if json.Unmarshal(data, &msg) == nil && msg.Type != "" {
		return msg
	}
	return message{Output: string(data)}
}

// send sends a control message.
func send(t *testing.T, ws *websocket.Conn, msg map[string]any) {
	t.Helper()
	data, err := json.Marshal(msg)
	assert.NoError(t, err)
	assert.NoError(t, websocket.Message.Send(ws, string(data)))
}

// untilStatus reads frames until the status, returning the output seen.
func untilStatus(t *testing.T, ws *websocket.Conn) (string, message) {
	t.Helper()
	var out strings.Builder
	for {
		msg := read(t, ws)
		if msg.Type == "status" {
			return out.String(), msg
		}
		out.WriteString(msg.Output)
	}
}

func TestAttach_Interactive(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	id := f.start(t, asOps, "infra", "echo")
	ws := f.attach(t, asOps, id)
	send(t, ws, map[string]any{"type": "resize", "cols": 100, "rows": 30})
	send(t, ws, map[string]any{"type": "input", "message": "hello\n"})

	out, status := untilStatus(t, ws)
	assert.Contains(t, out, "got hello")
	assert.Equal(t, model.StateOK, status.State)
	assert.Equal(t, 0, status.ExitCode)
}

func TestAttach_ReadOnly(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	id := f.start(t, asOps, "infra", "slow")
	ws := f.attach(t, asViewer, id)
	send(t, ws, map[string]any{"type": "input", "message": "ignored\n"})
	send(t, ws, map[string]any{"type": "stop"})
	send(t, ws, map[string]any{"type": "kill"})

	// The run ignores the viewer's ends and keeps running.
	time.Sleep(200 * time.Millisecond)
	res := f.call(t, asViewer, http.MethodGet, "/runs/"+id, nil)
	assert.Equal(t, model.StateRunning, res.str("state"))
	assert.Equal(t, float64(1), res.get("attached"))

	// A runner's stop ends it, and the viewer hears the status.
	send(t, f.attach(t, asOps, id), map[string]any{"type": "stop"})
	_, status := untilStatus(t, ws)
	assert.Equal(t, model.StateStopped, status.State)
}

func TestAttach_DetachKeepsRunning(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	id := f.start(t, asOps, "infra", "slow")
	ws := f.attach(t, asOps, id)
	assert.Eventually(t, func() bool {
		return f.call(t, asOps, http.MethodGet, "/runs/"+id, nil).get("attached") == float64(1)
	}, 5*time.Second, 20*time.Millisecond)
	ws.Close()

	assert.Eventually(t, func() bool {
		return f.call(t, asOps, http.MethodGet, "/runs/"+id, nil).get("attached") == float64(0)
	}, 5*time.Second, 20*time.Millisecond)
	assert.Equal(t, model.StateRunning, f.call(t, asOps, http.MethodGet, "/runs/"+id, nil).str("state"))

	// Attach again: still there, still running.
	again := f.attach(t, asOps, id)
	send(t, again, map[string]any{"type": "kill"})
	_, status := untilStatus(t, again)
	assert.Equal(t, model.StateKilled, status.State)
}

func TestAttach_Finished(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	run := &model.Run{Folder: "infra", Job: "plan", State: model.StateFailed, ExitCode: 2}
	assert.NoError(t, f.opts.Runs.Create(t.Context(), run))

	ws := f.attach(t, asOps, run.ID)
	msg := read(t, ws)
	assert.Equal(t, "status", msg.Type)
	assert.Equal(t, model.StateFailed, msg.State)
	assert.Equal(t, 2, msg.ExitCode)
	assert.Contains(t, msg.Message, "recording")
}

func TestAttach_Unknown(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	ws := f.attach(t, asOps, model.NewID())
	msg := read(t, ws)
	assert.Equal(t, "error", msg.State)
	assert.Contains(t, msg.Message, "not found")

	hidden := f.start(t, asAdmin, "private", "deploy")
	msg = read(t, f.attach(t, asLimited, hidden))
	assert.Equal(t, "error", msg.State)
}

func TestAttach_CrossOrigin(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	id := f.start(t, asOps, "infra", "slow")
	_, err := f.dial(t, asOps, id, "http://elsewhere.example")
	assert.Error(t, err)
}
