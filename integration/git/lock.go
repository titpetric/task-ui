package git

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// indexLock returns the checkout's index lock file, "" when there is none.
func indexLock(dir string) string {
	path := filepath.Join(dir, ".git", "index.lock")
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}

// lockHolder names the process holding a lock file, "" when none does and the lock is stale.
func lockHolder(ctx context.Context, path string) string {
	if pid := lockPID(path); pid > 0 {
		if alive(pid) {
			return "pid " + strconv.Itoa(pid)
		}
		return ""
	}
	fuser, err := exec.LookPath("fuser")
	if err != nil {
		return ""
	}
	ask, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ask, fuser, path).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return ""
	}
	return "pid " + strings.Join(strings.Fields(string(out)), " ")
}

// lockPID returns the process a lock file names in its first line, 0 when it names none.
func lockPID(path string) int {
	body, err := os.ReadFile(path)
	if err != nil || len(body) > 32 {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(body)))
	if err != nil {
		return 0
	}
	return pid
}

// alive reports whether a process of the pid is still running.
func alive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
