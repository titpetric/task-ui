//go:build integration

package agent_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/titpetric/task-ui/agent"
	"github.com/titpetric/task-ui/internal/assert"
)

// fakeClaude puts a claude first in a PATH that prints where it runs, the
// sandbox flag, HOME and its arguments, one per line, and exits with
// FAKE_EXIT. It returns the PATH.
func fakeClaude(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	body := "#!/bin/sh\necho \"cwd=$(pwd -P)\"\necho \"sandbox=$IS_SANDBOX\"\necho \"home=$HOME\"\nfor a; do echo \"arg=$a\"; done\nexit ${FAKE_EXIT:-0}\n"
	claude := filepath.Join(dir, "claude")
	assert.NoError(t, os.WriteFile(claude, []byte(body), 0o755))
	// A parallel test that forks while the file is open for writing hands
	// the child the open file until it execs, and running the script then
	// fails with "text file busy". Once one run starts, no child holds it.
	for i := 0; i < 100; i++ {
		if err := exec.Command(claude).Run(); !errors.Is(err, syscall.ETXTBSY) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	return dir + string(os.PathListSeparator) + os.Getenv("PATH")
}

// realDir resolves a temporary directory the way pwd -P prints it.
func realDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	assert.NoError(t, err)
	return dir
}

// TestMain runs the shell command in a process of its own under
// AGENT_TEST_SHELL: the command replaces its process with claude.
func TestMain(m *testing.M) {
	if args, ok := os.LookupEnv("AGENT_TEST_SHELL"); ok {
		a := &agent.Agent{Env: os.Environ(), Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr}
		if err := a.Run(context.Background(), append([]string{"shell"}, strings.Fields(args)...)); err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
			os.Exit(2)
		}
	}
	os.Exit(m.Run())
}
