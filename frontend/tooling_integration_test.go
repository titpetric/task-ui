//go:build integration

package frontend_test

import (
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
)

// stubAtkins writes an atkins that runs body and points the fixture's
// prober at it. The prober keeps no answer, so a later case replaces
// this one.
func stubAtkins(t *testing.T, f *fixture, body string) {
	t.Helper()

	command := filepath.Join(t.TempDir(), "atkins")
	assert.NoError(t, os.WriteFile(command, []byte("#!/bin/sh\n[ \"$1\" = warm ] && exit 0\n"+body+"\n"), 0o755))

	// A parallel test that forks while the file is open for writing
	// leaves the open file to the child until it execs: wait for one clean
	// run.
	for range 100 {
		if err := exec.Command(command, "warm").Run(); !errors.Is(err, syscall.ETXTBSY) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	f.opts.Tooling.Command = command
}

const toolingListing = `cat <<'JSON'
{"project":{"name":"project","repository":"titpetric/task-ui",
  "database":"/w/task-ui/.atkins/atkins.db","days":31,"from":"2026-09-22","to":"2026-10-23",
  "runs":44,"total_ms":184900,
  "commands":[{"command":"go test -cover","tool":"go","runs":42,"total_ms":184000,"average_ms":4380,"max_ms":9100,"failures":3}],
  "tools":[{"tool":"go","commands":1,"runs":42,"total_ms":184000,"failures":3}],
  "usage":[{"date":"2026-10-21","tool":"go","runs":20,"total_ms":90000},
           {"date":"2026-10-23","tool":"go","runs":22,"total_ms":94000},
           {"date":"2026-10-23","tool":"npm","runs":2,"total_ms":900}]},
 "workspace":{"name":"workspace",
  "database":"/home/you/.atkins/atkins.db","days":31,"from":"2026-09-22","to":"2026-10-23",
  "runs":310,"total_ms":900000,
  "commands":[{"command":"docker compose up","tool":"docker","runs":12,"total_ms":400000}],
  "tools":[{"tool":"docker","commands":1,"runs":12,"total_ms":400000}],
  "usage":[{"date":"2026-10-23","tool":"docker","runs":12,"total_ms":400000}]}}
JSON`

// The panel is on the runs page, drawn from what atkins recorded for
// the tree: a chart for each window, the tools behind them, and the
// commands behind the toggle.
func TestToolingOnTheRunsPage(t *testing.T) {
	f := newFixture(t)
	stubAtkins(t, f, toolingListing)

	res := f.get(t, asAdmin, "/runs")
	assert.Equal(t, http.StatusOK, res.Status)

	assert.Contains(t, res.Body, "Tooling")
	assert.Contains(t, res.Body, "Last 30 days")
	assert.Contains(t, res.Body, "Last 7 days")
	assert.Contains(t, res.Body, "2026-10-23: 24 commands in ")

	// A section per database atkins records to, the project named by
	// its repository and the workspace by nothing.
	assert.Contains(t, res.Body, "titpetric/task-ui")
	assert.Contains(t, res.Body, "/w/task-ui/.atkins/atkins.db")
	assert.Contains(t, res.Body, "/home/you/.atkins/atkins.db")
	assert.Contains(t, res.Body, "docker compose up")

	// The tools of a window, and the table behind the toggle.
	assert.Contains(t, res.Body, "tooling-legend")
	assert.Contains(t, res.Body, "<details")
	assert.Contains(t, res.Body, "go test -cover")
}

// An atkins that answers nothing leaves the page without the panel:
// the runs stand on their own.
func TestToolingWithoutAnAnswer(t *testing.T) {
	f := newFixture(t)

	// No atkins at all, the fixture's default.
	res := f.get(t, asAdmin, "/runs")
	assert.Equal(t, http.StatusOK, res.Status)
	assert.NotContains(t, res.Body, "tooling-charts")

	// One that fails, and one that has recorded nothing.
	stubAtkins(t, f, "exit 1")
	res = f.get(t, asAdmin, "/runs")
	assert.Equal(t, http.StatusOK, res.Status)
	assert.NotContains(t, res.Body, "tooling-charts")

	stubAtkins(t, f, `echo '{"workspace":{"name":"workspace","database":"/w/x.db","days":31,"to":"2026-10-23"}}'`)
	res = f.get(t, asAdmin, "/runs")
	assert.Equal(t, http.StatusOK, res.Status)
	assert.NotContains(t, res.Body, "tooling-charts")
}

// The panel reports commands and not runs, so it needs no permission of
// its own: a reader of the runs page sees it.
func TestToolingForEveryReaderOfTheRuns(t *testing.T) {
	f := newFixture(t)
	stubAtkins(t, f, toolingListing)

	for _, user := range []string{asAdmin, asOps, asViewer} {
		res := f.get(t, user, "/runs")
		assert.Equal(t, http.StatusOK, res.Status, user)
		assert.Contains(t, res.Body, "tooling-charts", user)
	}
}
