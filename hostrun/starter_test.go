package hostrun_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite" // the driver storage opens

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/hostrun"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/sshprobe"
	"github.com/titpetric/task-ui/storage"
)

// root is a catalogue of the root folder alone.
type root struct{ f *folder.Folder }

func (r root) Find(id string) *folder.Folder {
	if id == model.Default {
		return r.f
	}
	return nil
}

// starter is a starter over a hub of three hosts, an ssh that signs in
// to dev1 and dev2 as debian and runs the remote command locally, and a
// runner over a fresh database.
func starter(t *testing.T) (*hostrun.Starter, *storage.RunStorage) {
	t.Helper()
	hub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[` +
			`{"id":"a","name":"dev1","host":"dev1","status":"up","info":{"c":8}},` +
			`{"id":"b","name":"dev2","host":"dev2","status":"up","info":{"c":4}},` +
			`{"id":"c","name":"dev3","host":"dev3","status":"down","info":{"c":2}}]}`))
	}))
	t.Cleanup(hub.Close)
	ssh := filepath.Join(t.TempDir(), "ssh")
	body := "#!/bin/sh\nlogin=$(eval echo \\${$(($# - 1))})\nremote=$(eval echo \\${$#})\ncase \"$login\" in debian@dev1|debian@dev2) exec sh -c \"$remote\";; esac\nexit 255\n"
	assert.NoError(t, os.WriteFile(ssh, []byte(body), 0o755))

	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	runs := storage.NewRunStorage(db)
	manager := runner.New(runner.Options{Runs: runs, Artifacts: storage.NewArtifactStorage(db), Recordings: &recording.Store{Root: t.TempDir()}, ArtifactRoot: t.TempDir(), Grace: time.Second, Keep: time.Minute})
	t.Cleanup(func() { manager.Shutdown(context.Background()) })
	return &hostrun.Starter{
		Catalogue: root{&folder.Folder{ID: model.Default, Dir: t.TempDir()}},
		Hosts:     &beszel.Inventory{TTL: time.Minute, Settings: func(context.Context) (string, string, error) { return hub.URL, "t1", nil }},
		SSH:       &sshprobe.Prober{Users: []string{"debian"}, Timeout: 2 * time.Second, TTL: time.Hour, Command: ssh, Terminal: true},
		Runner:    manager,
		Timeout:   5 * time.Second,
	}, runs
}

// wait waits for the sessions to end.
func wait(t *testing.T, sessions []*runner.Session) {
	t.Helper()
	for _, session := range sessions {
		select {
		case <-session.Done():
		case <-time.After(10 * time.Second):
			t.Fatal("the run did not end")
		}
	}
}

// * runs on every host with a login, a batch; a named host runs alone,
// and one without a login fails saying so.
func TestStarter_Start(t *testing.T) {
	t.Parallel()
	s, runs := starter(t)
	ctx := context.Background()
	user := &model.User{ID: "u1", Name: "ops"}

	batch, sessions, err := s.Start(ctx, user, hostrun.Request{Targets: []string{"*"}, Command: "echo hi from $(echo here)"})
	assert.NoError(t, err)
	assert.True(t, model.ValidID(batch))
	assert.Len(t, sessions, 2)
	wait(t, sessions)
	rows, err := runs.List(ctx, storage.RunFilter{BatchID: batch})
	assert.NoError(t, err)
	assert.Len(t, rows, 2)
	for _, row := range rows {
		assert.Equal(t, model.StateOK, row.State)
		assert.Equal(t, model.HostJob(row.Target), row.Job)
		assert.Equal(t, model.Default, row.Folder)
		assert.Equal(t, []string{"echo hi from $(echo here)"}, row.ArgList())
		assert.NotEqual(t, "", row.CastPath)
	}

	alone, sessions, err := s.Start(ctx, user, hostrun.Request{Targets: []string{"dev3"}, Command: "uptime"})
	assert.NoError(t, err)
	assert.Equal(t, "", alone)
	wait(t, sessions)
	row, err := runs.Get(ctx, sessions[0].ID())
	assert.NoError(t, err)
	assert.Equal(t, "dev3", row.Target)
	assert.Equal(t, model.StateFailed, row.State)

	// ssh that can't run the command says so, beside its exit code.
	_, sessions, err = s.Start(ctx, user, hostrun.Request{Targets: []string{"dev1"}, Command: "exit 255"})
	assert.NoError(t, err)
	out, status := collect(t, sessions[0])
	assert.Equal(t, 255, status.ExitCode)
	assert.Contains(t, out, "ssh to debian@dev1 ended with 255")

	_, _, err = s.Start(ctx, user, hostrun.Request{Targets: []string{"elsewhere"}, Command: "uptime"})
	assert.ErrorIs(t, err, hostrun.ErrUnknownHost)
	_, _, err = s.Start(ctx, user, hostrun.Request{Command: " "})
	assert.ErrorIs(t, err, hostrun.ErrNoCommand)
}

// Resolve names every host with a login for *, in the inventory's
// order, a named host whatever its login, and refuses one it doesn't
// list.
func TestStarter_Resolve(t *testing.T) {
	t.Parallel()
	s, _ := starter(t)
	ctx := context.Background()
	hosts, logins, err := s.Resolve(ctx, []string{"*"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"dev1", "dev2"}, hosts)
	assert.Equal(t, "debian@dev1", logins["dev1"])
	hosts, _, err = s.Resolve(ctx, []string{"dev3"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"dev3"}, hosts)
	_, _, err = s.Resolve(ctx, []string{"elsewhere"})
	assert.ErrorIs(t, err, hostrun.ErrUnknownHost)
}

// collect reads a session until its status, returning the output seen.
func collect(t *testing.T, session *runner.Session) (string, runner.Status) {
	t.Helper()
	sub := session.Attach(true)
	var out strings.Builder
	deadline := time.After(10 * time.Second)
	for {
		select {
		case event, ok := <-sub.C:
			if !ok {
				t.Fatalf("closed without a status, output %q", out.String())
			}
			out.Write(event.Output)
			if event.Status != nil {
				return out.String(), *event.Status
			}
		case <-deadline:
			t.Fatalf("no status after 10s, output %q", out.String())
		}
	}
}

// A second command on a host still running one waits, queued, and runs
// once the first ended; the other hosts don't wait.
func TestStarter_Queue(t *testing.T) {
	t.Parallel()
	s, runs := starter(t)
	ctx := context.Background()
	user := &model.User{ID: "u1", Name: "ops"}
	gate := filepath.Join(t.TempDir(), "gate")

	_, first, err := s.Start(ctx, user, hostrun.Request{Targets: []string{"dev1"}, Command: `while [ ! -f "` + gate + `" ]; do sleep 0.05; done`})
	assert.NoError(t, err)
	_, second, err := s.Start(ctx, user, hostrun.Request{Targets: []string{"dev1"}, Command: "echo after"})
	assert.NoError(t, err)
	assert.Equal(t, model.StateQueued, second[0].Status().State)
	_, other, err := s.Start(ctx, user, hostrun.Request{Targets: []string{"dev2"}, Command: "true"})
	assert.NoError(t, err)
	wait(t, other)
	assert.Equal(t, model.StateQueued, second[0].Status().State, "dev2's run doesn't start dev1's queue")

	assert.NoError(t, os.WriteFile(gate, nil, 0o644))
	wait(t, append(first, second...))
	row, err := runs.Get(ctx, second[0].ID())
	assert.NoError(t, err)
	assert.Equal(t, model.StateOK, row.State)
	assert.Equal(t, "@host:dev1", row.Job)
	assert.True(t, row.StartedAt.After(*first[0].Snapshot().StartedAt), "the start is when it left the queue")
}
