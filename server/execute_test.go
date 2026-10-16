package server

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/secret"
)

func TestExecute(t *testing.T) {
	ts, svc := testServer(t)
	ops := login(t, ts, "ops")

	ws, err := dial(t, ts, ops, "/ws/run/echo/hello?cols=100&rows=20", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	out, statuses := collect(t, ws, final)
	ws.Close()

	if !strings.Contains(out, "hello") {
		t.Fatalf("output: %q", out)
	}
	last := statuses[len(statuses)-1]
	if statuses[0].State != "running" || last.State != "ok" || last.Exit == nil || *last.Exit != 0 {
		t.Fatalf("statuses: %+v", statuses)
	}
	records := svc.history.List("echo", "hello")
	if len(records) != 1 || records[0].User != "ops" || records[0].State != "ok" {
		t.Fatalf("history: %+v", records)
	}
}

func TestExecuteSecrets(t *testing.T) {
	ts, svc := testServer(t)
	ops := login(t, ts, "ops")

	// A shared secret, a folder secret that wins, and a managed env file.
	if err := svc.secrets.Set(secret.Secret{Scope: secret.ScopeAll, Key: "GREETING", Value: "hello from all"}); err != nil {
		t.Fatal(err)
	}
	if err := svc.secrets.Set(secret.Secret{Scope: "printenv", Key: "GREETING", Value: "hello from the folder"}); err != nil {
		t.Fatal(err)
	}
	f := svc.folders().Find("printenv")
	if err := svc.env.Save(f, envfile.File{Path: ".env", Content: "EDITED=yes\n"}); err != nil {
		t.Fatal(err)
	}

	ws, err := dial(t, ts, ops, "/ws/run/printenv/GREETING", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	out, statuses := collect(t, ws, final)
	ws.Close()
	if !strings.Contains(out, "hello from the folder") || statuses[len(statuses)-1].State != "ok" {
		t.Fatalf("run output %q, statuses %+v", out, statuses)
	}
	data, err := os.ReadFile(filepath.Join(f.Dir, ".env"))
	if err != nil || !strings.Contains(string(data), "EDITED=yes") || !strings.Contains(string(data), "GREETING=\"hello from the folder\"") {
		t.Fatalf("generated env file: %q %v", data, err)
	}

	// Deleting the folder secret falls back to the shared one.
	if err := svc.secrets.Delete(secret.Secret{Scope: "printenv", Key: "GREETING"}); err != nil {
		t.Fatal(err)
	}
	ws, err = dial(t, ts, ops, "/ws/run/printenv/GREETING", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	out, _ = collect(t, ws, final)
	ws.Close()
	if !strings.Contains(out, "hello from all") {
		t.Fatalf("shared secret not applied: %q", out)
	}
}

func TestExecuteInteractive(t *testing.T) {
	ts, _ := testServer(t)
	c := login(t, ts, "ops")

	ws, err := dial(t, ts, c, "/ws/run/cat/-", ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	if _, statuses := collect(t, ws, running); len(statuses) == 0 {
		t.Fatal("run did not start")
	}

	// Typed input is a control frame; cat echoes it through the pty. A
	// resize is accepted along the way, and a frame that is not a control
	// message is ignored.
	send(t, ws, `{"type":"resize","cols":90,"rows":25}`)
	send(t, ws, `not json`)
	send(t, ws, `{"type":"input","message":"hi there\n"}`)
	readOutput(t, ws, "hi there")

	// Stop ends the run as stopped.
	send(t, ws, `{"type":"stop"}`)
	_, statuses := collect(t, ws, final)
	if len(statuses) == 0 || statuses[len(statuses)-1].State != "stopped" {
		t.Fatalf("after stop: %+v", statuses)
	}
}

func TestCrlfWriter(t *testing.T) {
	var out bytes.Buffer
	n, err := crlfWriter{&out}.Write([]byte("a\nb\n"))
	if err != nil || n != 4 || out.String() != "a\r\nb\r\n" {
		t.Fatalf("crlf: %d %v %q", n, err, out.String())
	}
}
