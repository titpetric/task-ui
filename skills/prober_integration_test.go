//go:build integration

package skills_test

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
	"github.com/titpetric/task-ui/skills"
)

// fakeAtkins writes an atkins that logs each call to calls, refuses a
// pipe on stdin, and prints body with $PWD in place of DIR.
func fakeAtkins(t *testing.T, body string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	calls := filepath.Join(dir, "calls")
	script := "#!/bin/sh\n" +
		"[ \"$1\" = warm ] && exit 0\n" +
		"echo \"$PWD $*\" >> " + calls + "\n" +
		"if [ -p /dev/stdin ]; then echo piped >&2; exit 3; fi\n" +
		body + "\n"
	command := filepath.Join(dir, "atkins")
	assert.NoError(t, os.WriteFile(command, []byte(script), 0o755))
	// A parallel test that forks while the file is open for writing hands
	// the child the open file until it execs: wait for one clean run.
	for i := 0; i < 100; i++ {
		if err := exec.Command(command, "warm").Run(); !errors.Is(err, syscall.ETXTBSY) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	return command, calls
}

func lines(t *testing.T, file string) []string {
	t.Helper()
	data, err := os.ReadFile(file)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	assert.NoError(t, err)
	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

const listing = `cat <<JSON
[{"id":"","kind":"pipeline","path":"$PWD/atkins.yml","dir":"$PWD","active":true,"cmds":[{"id":"build","name":"build","cmd":"atkins build","active":true}]},
 {"id":"git","kind":"skill","path":"/ws/.atkins/skills/git.yml","dir":"/ws","active":true,"when":{"files":[".git"],"matched":"/ws/.git"},"cmds":[{"id":"git:status","name":"status","desc":"Status","cmd":"atkins git:status","active":true}]},
 {"id":"go","kind":"skill","path":"/ws/.atkins/skills/go.yml","active":false,"cmds":[]}]
JSON`

// Probe runs atkins -l --json --all in the directory with nothing on
// stdin, reads its sections, and keeps them until the TTL passes or a
// file it read changes.
func TestProber_Probe(t *testing.T) {
	t.Parallel()
	command, calls := fakeAtkins(t, listing)
	dir, err := filepath.EvalSymlinks(t.TempDir())
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "atkins.yml"), []byte("jobs: {}\n"), 0o644))
	p := &skills.Prober{Command: command, TTL: time.Hour}

	sections := p.Probe(dir)
	assert.Len(t, sections, 3)
	assert.Equal(t, skills.KindPipeline, sections[0].Kind)
	assert.Equal(t, dir, sections[0].Dir)
	assert.Equal(t, "git:status", sections[1].Cmds[0].ID)
	assert.Equal(t, "/ws/.git", sections[1].When.Matched)
	assert.Equal(t, []string{dir + " -l --json --all"}, lines(t, calls))

	p.Probe(dir)
	assert.Len(t, lines(t, calls), 1, "cached")

	later := time.Now().Add(time.Minute)
	assert.NoError(t, os.Chtimes(filepath.Join(dir, "atkins.yml"), later, later))
	p.Probe(dir)
	assert.Len(t, lines(t, calls), 2, "the pipeline changed")

	skillsDir := filepath.Join(dir, ".atkins", "skills")
	assert.NoError(t, os.MkdirAll(skillsDir, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(skillsDir, "go.yml"), []byte("jobs: {}\n"), 0o644))
	p.Probe(dir)
	assert.Len(t, lines(t, calls), 3, "a skill appeared")

	expired := &skills.Prober{Command: command, TTL: time.Nanosecond}
	expired.Probe(dir)
	time.Sleep(time.Millisecond)
	expired.Probe(dir)
	assert.Len(t, lines(t, calls), 5, "past the TTL")
}

// A probe that fails, prints something other than the listing, or runs
// past the timeout answers nothing; so does a missing atkins.
func TestProber_Fails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for name, body := range map[string]string{
		"exit":  "echo '{\"error\":\"lint\"}'; exit 1",
		"json":  "echo '{\"error\":\"lint\"}'",
		"slow":  "sleep 5",
		"plain": "echo nothing",
	} {
		command, _ := fakeAtkins(t, body)
		p := &skills.Prober{Command: command, Timeout: 200 * time.Millisecond}
		start := time.Now()
		assert.Nil(t, p.Probe(dir), name)
		assert.Less(t, time.Since(start), 3*time.Second, name)
	}
	assert.Nil(t, (&skills.Prober{Command: filepath.Join(dir, "missing")}).Probe(dir))
	assert.Nil(t, (*skills.Prober)(nil).Probe(dir))
	assert.Nil(t, (&skills.Prober{}).Probe(""))
}
