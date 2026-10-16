package run

import (
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestProcess(t *testing.T) {
	proc, err := Start(exec.Command("sh", "-c", "stty size; echo hello"), 100, 40)
	if err != nil {
		t.Fatal(err)
	}
	defer proc.TTY.Close()
	out, _ := io.ReadAll(proc.TTY)
	if code := proc.Wait(); code != 0 {
		t.Fatalf("exit: %d", code)
	}
	if text := string(out); !strings.Contains(text, "40 100") || !strings.Contains(text, "hello") {
		t.Fatalf("pty output: %q", text)
	}
}

func TestProcess_Resize(t *testing.T) {
	proc, err := Start(exec.Command("sh", "-c", "sleep 0.2; stty size"), 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer proc.TTY.Close()
	proc.Resize(132, 50)
	proc.Resize(0, 0) // ignored
	out, _ := io.ReadAll(proc.TTY)
	proc.Wait()
	if !strings.Contains(string(out), "50 132") {
		t.Fatalf("resized size: %q", out)
	}
}

func TestProcess_Stop(t *testing.T) {
	// The child starts a grandchild; stopping the group ends both.
	proc, err := Start(exec.Command("sh", "-c", "sleep 30 & wait"), 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer proc.TTY.Close()
	go io.Copy(io.Discard, proc.TTY)
	time.Sleep(50 * time.Millisecond)
	proc.Stop(time.Second)
	done := make(chan int, 1)
	go func() { done <- proc.Wait() }()
	select {
	case code := <-done:
		if code == 0 {
			t.Fatal("a stopped process reported success")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Stop did not end the process group")
	}
	// Stopping again after exit is a no-op.
	proc.Stop(time.Second)
}

func TestProcess_Wait(t *testing.T) {
	proc, err := Start(exec.Command("sh", "-c", "exit 3"), 80, 24)
	if err != nil {
		t.Fatal(err)
	}
	defer proc.TTY.Close()
	io.Copy(io.Discard, proc.TTY)
	if code := proc.Wait(); code != 3 {
		t.Fatalf("exit code: %d", code)
	}
	if _, err := Start(exec.Command("/nonexistent/binary"), 80, 24); err == nil {
		t.Fatal("missing binary started")
	}
}
