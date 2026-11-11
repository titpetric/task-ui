package git

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestIndexLock(t *testing.T) {
	dir := t.TempDir()
	if got := indexLock(dir); got != "" {
		t.Errorf("a folder with no .git has the lock %q", got)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, ".git", "index.lock")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := indexLock(dir); got != path {
		t.Errorf("the lock is %q, want %q", got, path)
	}
}

func TestLockPID(t *testing.T) {
	dir := t.TempDir()
	for _, test := range []struct {
		body string
		want int
	}{
		{"", 0},
		{strconv.Itoa(os.Getpid()) + "\n", os.Getpid()},
		{"DIRC\x00\x00\x00\x02", 0},
	} {
		path := filepath.Join(dir, "lock")
		if err := os.WriteFile(path, []byte(test.body), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := lockPID(path); got != test.want {
			t.Errorf("%q names pid %d, want %d", test.body, got, test.want)
		}
	}
}

func TestAlive(t *testing.T) {
	if !alive(os.Getpid()) {
		t.Error("this process is not alive")
	}
	if alive(1 << 24) {
		t.Error("a pid nothing runs under is alive")
	}
}
