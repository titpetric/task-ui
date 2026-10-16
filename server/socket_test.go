package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

// dial opens the run socket as the client's user.
func dial(t *testing.T, ts *httptest.Server, c *http.Client, path, origin string) (*websocket.Conn, error) {
	t.Helper()
	u, _ := url.Parse(ts.URL)
	config, err := websocket.NewConfig("ws://"+u.Host+path, origin)
	if err != nil {
		t.Fatal(err)
	}
	for _, cookie := range c.Jar.Cookies(u) {
		config.Header.Add("Cookie", cookie.Name+"="+cookie.Value)
	}
	return websocket.DialConfig(config)
}

// send writes one control frame.
func send(t *testing.T, ws *websocket.Conn, frame string) {
	t.Helper()
	if err := websocket.Message.Send(ws, frame); err != nil {
		t.Fatal(err)
	}
}

// collect reads frames until a status satisfies until, returning the
// terminal output and the statuses seen.
func collect(t *testing.T, ws *websocket.Conn, until func(control) bool) (string, []control) {
	t.Helper()
	var out strings.Builder
	var statuses []control
	if err := ws.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		t.Fatal(err)
	}
	for {
		var frame []byte
		if err := websocket.Message.Receive(ws, &frame); err != nil {
			return out.String(), statuses
		}
		if len(frame) > 0 && frame[0] == '{' {
			var msg control
			if json.Unmarshal(frame, &msg) == nil && msg.Type == "status" {
				statuses = append(statuses, msg)
				if until(msg) {
					return out.String(), statuses
				}
				continue
			}
		}
		out.Write(frame)
	}
}

// readOutput reads terminal frames until the output holds want.
func readOutput(t *testing.T, ws *websocket.Conn, want string) string {
	t.Helper()
	var out strings.Builder
	if err := ws.SetReadDeadline(time.Now().Add(10 * time.Second)); err != nil {
		t.Fatal(err)
	}
	for !strings.Contains(out.String(), want) {
		var frame []byte
		if err := websocket.Message.Receive(ws, &frame); err != nil {
			t.Fatalf("waiting for %q: %v, got %q", want, err, out.String())
		}
		if len(frame) > 0 && frame[0] != '{' {
			out.Write(frame)
		}
	}
	return out.String()
}

func final(msg control) bool { return msg.State != "running" }

func running(msg control) bool { return msg.State == "running" }

func TestRunSocket(t *testing.T) {
	ts, _ := testServer(t)
	c := login(t, ts, "admin")
	if _, err := dial(t, ts, c, "/ws/run/echo/hello", "http://evil.example"); err == nil {
		t.Fatal("cross origin websocket accepted")
	}
	ws, err := dial(t, ts, c, "/ws/run/echo/nothing", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	if _, statuses := collect(t, ws, final); len(statuses) != 1 || statuses[0].State != "error" {
		t.Fatalf("missing job: %+v", statuses)
	}
}

func TestServeRunDenied(t *testing.T) {
	ts, _ := testServer(t)

	guest := login(t, ts, "guest")
	ws, err := dial(t, ts, guest, "/ws/run/echo/hello", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, statuses := collect(t, ws, final)
	ws.Close()
	if len(statuses) != 1 || statuses[0].State != "denied" {
		t.Fatalf("viewer run: %+v", statuses)
	}

	// A runner is held to the allow list of the job.
	ops := login(t, ts, "ops")
	ws, err = dial(t, ts, ops, "/ws/run/echo/locked", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, statuses = collect(t, ws, final)
	ws.Close()
	if len(statuses) != 1 || statuses[0].State != "denied" {
		t.Fatalf("allow list ignored: %+v", statuses)
	}
}

func TestServeRunLock(t *testing.T) {
	ts, svc := testServer(t)
	alice := login(t, ts, "ops")
	bob := login(t, ts, "admin")

	first, err := dial(t, ts, alice, "/ws/run/sleep/5", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	if _, statuses := collect(t, first, running); len(statuses) == 0 {
		t.Fatal("first run did not start")
	}

	// A second run of the same job is refused.
	second, err := dial(t, ts, bob, "/ws/run/sleep/5", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, statuses := collect(t, second, final)
	second.Close()
	if len(statuses) != 1 || statuses[0].State != "busy" || statuses[0].User != "ops" {
		t.Fatalf("second run: %+v", statuses)
	}

	// Run force kills the first run and takes over.
	forced, err := dial(t, ts, bob, "/ws/run/sleep/5?force=1&cols=100&rows=30", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	if _, statuses := collect(t, forced, running); len(statuses) == 0 || statuses[len(statuses)-1].State != "running" {
		t.Fatalf("forced run: %+v", statuses)
	}
	if _, statuses := collect(t, first, final); len(statuses) == 0 || statuses[len(statuses)-1].State != "killed" {
		t.Fatalf("first run after force: %+v", statuses)
	}

	// Closing the window stops the run and frees the lock.
	forced.Close()
	waitFor(t, func() bool {
		_, busy := svc.runs.Get("sleep/5")
		return !busy
	})
	waitFor(t, func() bool {
		last, ok := svc.history.Last("sleep", "5")
		return ok && last.State == "stopped" && last.Forced
	})
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("condition not met")
		}
		time.Sleep(20 * time.Millisecond)
	}
}
