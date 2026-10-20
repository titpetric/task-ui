package sshprobe

import (
	"time"
)

// Result is how a command went on one host.
type Result struct {
	Host     string `json:"host"`
	Login    string `json:"login"`
	ExitCode int    `json:"exit_code"`
	// Output is what the command printed, both streams, cut at 256 KB.
	Output string `json:"output"`
	// Error says why the command did not run or end: no login, a lost
	// connection, the time limit.
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}
