package tooling_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/tooling"
)

// stub writes an atkins that prints what it is given and logs the
// directories it was asked about, and returns a prober over it with the
// log's path.
func stub(t *testing.T, body string) (*tooling.Prober, string) {
	t.Helper()

	command := filepath.Join(t.TempDir(), "atkins")
	script := "#!/bin/sh\n[ \"$1\" = warm ] && exit 0\npwd -P >> \"$0.calls\"\n" + body + "\n"
	assert.NoError(t, os.WriteFile(command, []byte(script), 0o755))

	// A parallel test that forks while the file is open for writing
	// hands the child the open file until it execs: wait for one clean
	// run.
	for range 100 {
		if err := exec.Command(command, "warm").Run(); !errors.Is(err, syscall.ETXTBSY) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	return &tooling.Prober{Command: command, TTL: time.Hour}, command + ".calls"
}

// calls returns the directories the stub was asked about.
func calls(t *testing.T, path string) []string {
	t.Helper()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	assert.NoError(t, err)
	return strings.Fields(string(data))
}

const listing = `cat <<'JSON'
{"database":"/w/.atkins/atkins.db","days":31,"from":"2026-09-22","to":"2026-10-23",
 "runs":44,"total_ms":184900,
 "commands":[{"command":"go test","tool":"go","runs":42,"total_ms":184000,"average_ms":4380,"max_ms":9100,"failures":3}],
 "tools":[{"tool":"go","commands":1,"runs":42,"total_ms":184000,"failures":3}],
 "usage":[{"date":"2026-10-23","tool":"go","runs":42,"total_ms":184000}]}
JSON`

func TestProber_Probe(t *testing.T) {
	prober, log := stub(t, listing)
	dir := t.TempDir()

	report := prober.Probe(dir)
	assert.NotNil(t, report)
	assert.Equal(t, "/w/.atkins/atkins.db", report.Database)
	assert.Equal(t, 31, report.Days)
	assert.Equal(t, int64(44), report.Runs)
	assert.Len(t, report.Commands, 1)
	assert.Equal(t, "go test", report.Commands[0].Command)
	assert.Len(t, report.Tools, 1)
	assert.Len(t, report.Usage, 1)

	// atkins is asked in the directory it is asked about, because that
	// is how it finds the database.
	assert.Equal(t, []string{dir}, calls(t, log))
}

// An answer is kept for the TTL, so a page anyone may reload does not
// run atkins once a reader.
func TestProber_ProbeCaches(t *testing.T) {
	prober, log := stub(t, listing)
	dir := t.TempDir()

	first := prober.Probe(dir)
	second := prober.Probe(dir)
	assert.NotNil(t, first)
	assert.Equal(t, first, second)
	assert.Len(t, calls(t, log), 1)

	// A different directory is a different question.
	other := t.TempDir()
	assert.NotNil(t, prober.Probe(other))
	assert.Equal(t, []string{dir, other}, calls(t, log))
}

// An atkins that fails, times out, or prints something that is not a
// report answers nothing, and the page shows no panel rather than
// refusing to load.
func TestProber_ProbeWithoutAnAnswer(t *testing.T) {
	failing, _ := stub(t, "exit 1")
	assert.Nil(t, failing.Probe(t.TempDir()))

	garbage, _ := stub(t, "echo not json")
	assert.Nil(t, garbage.Probe(t.TempDir()))

	missing := &tooling.Prober{Command: filepath.Join(t.TempDir(), "nothing"), TTL: time.Hour}
	assert.Nil(t, missing.Probe(t.TempDir()))

	slow := &tooling.Prober{Command: "sleep", Timeout: 10 * time.Millisecond, TTL: time.Hour}
	assert.Nil(t, slow.Probe(t.TempDir()))
}

// A prober nobody made, and a directory nobody named, answer nothing.
func TestProber_ProbeWithoutADirectory(t *testing.T) {
	var none *tooling.Prober
	assert.Nil(t, none.Probe("/w"))

	prober, log := stub(t, listing)
	assert.Nil(t, prober.Probe(""))
	assert.Len(t, calls(t, log), 0)
}

func TestProber_Bin(t *testing.T) {
	var none *tooling.Prober
	assert.Equal(t, "atkins", none.Bin())
	assert.Equal(t, "atkins", (&tooling.Prober{}).Bin())
	assert.Equal(t, "/usr/local/bin/atkins", (&tooling.Prober{Command: "/usr/local/bin/atkins"}).Bin())
}
