package history

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	s := &Store{Root: t.TempDir()}

	start := time.Now().Add(-time.Minute)
	rec, err := s.Start(Record{Folder: "infra", Job: "deploy", User: "ops", Started: start}, 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	if !ValidID(rec.Record.ID) {
		t.Fatalf("invalid generated id %q", rec.Record.ID)
	}

	// A run in progress is listed as running.
	list := s.List("infra", "deploy")
	if len(list) != 1 || list[0].State != StateRunning {
		t.Fatalf("in progress: %+v", list)
	}

	rec.Write([]byte("hello\r\n"))
	if err := rec.Finish(StateFailed, 2); err != nil {
		t.Fatal(err)
	}

	last, ok := s.Last("infra", "deploy")
	if !ok || last.State != StateFailed || last.ExitCode != 2 || last.User != "ops" {
		t.Fatalf("last: %+v", last)
	}
	if _, ok := s.Last("infra", "nothing"); ok {
		t.Fatal("last of a job never run")
	}

	path, err := s.CastPath(last)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 || !strings.Contains(lines[0], `"version":2`) || !strings.Contains(lines[1], `"hello\r\n"`) {
		t.Fatalf("cast: %q", data)
	}
	if _, err := s.CastPath(Record{Folder: "infra", Job: "deploy", ID: "../../etc/passwd"}); err == nil {
		t.Fatal("accepted a path as id")
	}

	second, err := s.Start(Record{Folder: "web", Job: "make:build", User: "a", Started: time.Now()}, 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	if err := second.Finish(StateOK, 0); err != nil {
		t.Fatal(err)
	}
	recent := s.Recent(10)
	if len(recent) != 2 || recent[0].Folder != "web" || recent[0].Job != "make:build" {
		t.Fatalf("recent: %+v", recent)
	}
	if len(s.Recent(1)) != 1 {
		t.Fatal("limit ignored")
	}
}
