package repository

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/server/model"
)

func TestHistoryRecord(t *testing.T) {
	h := &History{Root: t.TempDir()}

	start := time.Now().Add(-time.Minute)
	rec, err := h.Start(model.HistoryRecord{Project: "infra", Job: "deploy", User: "ops", Started: start}, 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	if !ValidID(rec.Record.ID) {
		t.Fatalf("invalid generated id %q", rec.Record.ID)
	}

	// A run in progress is listed as running.
	list := h.List("infra", "deploy")
	if len(list) != 1 || list[0].State != model.StateRunning {
		t.Fatalf("in progress: %+v", list)
	}

	rec.Write([]byte("hello\r\n"))
	if err := rec.Finish(model.StateFailed, 2); err != nil {
		t.Fatal(err)
	}

	last, ok := h.Last("infra", "deploy")
	if !ok || last.State != model.StateFailed || last.ExitCode != 2 || last.User != "ops" {
		t.Fatalf("last: %+v", last)
	}

	path, err := h.CastPath("infra", "deploy", last.ID)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 || !strings.Contains(lines[0], `"version":2`) || !strings.Contains(lines[1], `"hello\r\n"`) {
		t.Fatalf("cast: %q", data)
	}

	if _, err := h.CastPath("infra", "deploy", "../../etc/passwd"); err == nil {
		t.Fatal("accepted a path as id")
	}

	second, _ := h.Start(model.HistoryRecord{Project: "web", Job: "build", User: "a", Started: time.Now()}, 80, 24)
	second.Finish(model.StateOK, 0)
	recent := h.Recent(10)
	if len(recent) != 2 || recent[0].Project != "web" {
		t.Fatalf("recent: %+v", recent)
	}
}
