// Package skillstest is a stand-in atkins for the tests of the packages
// that probe skills: it lists a git skill in the nearest checkout and a
// compose skill beside a compose.yml, as the workspace skills do.
package skillstest

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/titpetric/task-ui/skills"
)

// script lists, for the directory it runs in: git (status, pull) when a
// .git is there or above, with that directory; compose (build only when
// the file builds, up, down, custom with args) when compose.yml is there;
// go, never active. It appends each directory it probes to $0.calls.
// Asked to run a job, it prints the job and where it runs.
const script = `#!/bin/sh
[ "$1" = warm ] && exit 0
here=$(pwd -P)
if [ "$1" != -l ]; then echo "ran $* in $here"; exit 0; fi
echo "$here" >> "$0.calls"
up=$here
while [ ! -e "$up/.git" ] && [ "$up" != / ]; do up=$(dirname "$up"); done
printf '['
sep=''
if [ -e "$up/.git" ]; then
  printf '{"id":"git","kind":"skill","path":"/ws/.atkins/skills/git.yml","dir":"%s","active":true,"when":{"files":[".git"],"matched":"%s/.git"},"desc":"Git","cmds":[{"id":"git:pull","name":"pull","desc":"Pull every checkout","cmd":"atkins git:pull","active":true},{"id":"git:status","name":"status","desc":"Show the branch","cmd":"atkins git:status","active":true}]}' "$up" "$up"
  sep=,
fi
if [ -f "$here/compose.yml" ]; then
  build=false
  grep -q 'build:' "$here/compose.yml" && build=true
  printf '%s{"id":"compose","kind":"skill","path":"/ws/.atkins/skills/compose.yml","dir":"%s","active":true,"when":{"files":["compose.yml"],"here":true,"matched":"%s/compose.yml"},"desc":"Docker Compose","cmds":[{"id":"compose:build","name":"build","cmd":"atkins compose:build","active":%s},{"id":"compose:custom","name":"custom","cmd":"atkins compose:custom","args":true,"active":true},{"id":"compose:down","name":"down","cmd":"atkins compose:down","active":true},{"id":"compose:up","name":"up","cmd":"atkins compose:up","active":true}]}' "$sep" "$here" "$here" "$build"
  sep=,
fi
printf '%s{"id":"go","kind":"skill","path":"/ws/.atkins/skills/go.yml","active":false,"cmds":[{"id":"go:test","name":"test","cmd":"atkins go:test","active":false}]}]\n' "$sep"
`

// Prober returns a Prober over the stand-in atkins, and the file it logs
// the directories it probes to.
func Prober(t testing.TB) (*skills.Prober, string) {
	t.Helper()
	command := filepath.Join(t.TempDir(), "atkins")
	if err := os.WriteFile(command, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	// A parallel test that forks while the file is open for writing hands
	// the child the open file until it execs: wait for one clean run.
	for i := 0; i < 100; i++ {
		if err := exec.Command(command, "warm").Run(); !errors.Is(err, syscall.ETXTBSY) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	return &skills.Prober{Command: command, TTL: time.Hour}, command + ".calls"
}
