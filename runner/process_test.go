package runner_test

import (
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

func TestStartProcess(t *testing.T) {
	t.Parallel()
	proc, err := runner.StartProcess(exec.Command("sh", "-c", "stty size; echo hello"), 100, 40)
	assert.NoError(t, err)
	assert.Positive(t, proc.PID())
	out, _ := io.ReadAll(proc)
	assert.Equal(t, 0, proc.Wait())
	assert.NoError(t, proc.Close())
	assert.Contains(t, string(out), "40 100")
	assert.Contains(t, string(out), "hello")

	_, err = runner.StartProcess(exec.Command("/nonexistent/binary"), 80, 24)
	assert.Error(t, err)
}

func TestProcess_Write(t *testing.T) {
	t.Parallel()
	proc, err := runner.StartProcess(exec.Command("cat"), 80, 24)
	assert.NoError(t, err)
	defer proc.Close()
	_, err = proc.Write([]byte("echoed\n"))
	assert.NoError(t, err)
	buf := make([]byte, 64)
	n, err := proc.Read(buf)
	assert.NoError(t, err)
	assert.Contains(t, string(buf[:n]), "echoed")
	proc.Stop(time.Second)
	assert.NotEqual(t, 0, proc.Wait())
}

func TestProcess_Resize(t *testing.T) {
	t.Parallel()
	proc, err := runner.StartProcess(exec.Command("sh", "-c", "sleep 0.2; stty size"), 80, 24)
	assert.NoError(t, err)
	defer proc.Close()
	proc.Resize(132, 50)
	proc.Resize(0, 0)
	out, _ := io.ReadAll(proc)
	proc.Wait()
	assert.Contains(t, string(out), "50 132")
}

func TestProcess_Stop(t *testing.T) {
	t.Parallel()
	// The child starts a grandchild; stopping the group ends both.
	proc, err := runner.StartProcess(exec.Command("sh", "-c", "sleep 30 & wait"), 80, 24)
	assert.NoError(t, err)
	defer proc.Close()
	go io.Copy(io.Discard, proc)
	time.Sleep(50 * time.Millisecond)
	proc.Stop(time.Second)
	done := make(chan int, 1)
	go func() { done <- proc.Wait() }()
	select {
	case code := <-done:
		assert.NotEqual(t, 0, code)
	case <-time.After(5 * time.Second):
		t.Fatal("Stop did not end the process group")
	}
	proc.Stop(time.Second)
}

func TestProcess_Wait(t *testing.T) {
	t.Parallel()
	proc, err := runner.StartProcess(exec.Command("sh", "-c", "exit 3"), 80, 24)
	assert.NoError(t, err)
	defer proc.Close()
	io.Copy(io.Discard, proc)
	assert.Equal(t, 3, proc.Wait())
	_ = strings.TrimSpace
}
