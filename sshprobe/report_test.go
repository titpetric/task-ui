package sshprobe_test

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshprobe"
)

// A report is a section a host, the output plain: no colour codes.
func TestReport(t *testing.T) {
	t.Parallel()
	report := sshprobe.Report("df -h", []sshprobe.Result{
		{Host: "a", Login: "debian@a", Output: "\x1b[01;34m/dev/sda1\x1b[0m 40%\n", Duration: time.Second},
		{Host: "b", ExitCode: -1, Error: "task-ui has no ssh login to this host"},
	})
	assert.Contains(t, report, "# `df -h` on 2 hosts")
	assert.Contains(t, report, "## debian@a\n\nexit 0, 1s\n\n```\n/dev/sda1 40%\n```")
	assert.Contains(t, report, "## b\n\nexit -1, 0s\n\nerror: task-ui has no ssh login")
	assert.NotContains(t, report, "\x1b")
}
