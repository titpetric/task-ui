package tooling

import (
	"time"
)

// Sample is one command atkins recorded, as `--stats --all` lists it:
// when it started, its key, how long it took and how it ended. It
// names the command by its key and never by its command line, which
// can carry a token.
type Sample struct {
	At           time.Time `json:"at"`
	Command      string    `json:"command"`
	Tool         string    `json:"tool"`
	Kind         string    `json:"kind"`
	DurationMs   int64     `json:"duration_ms"`
	ExitCode     int64     `json:"exit_code"`
	Failed       bool      `json:"failed"`
	Pipeline     string    `json:"pipeline,omitempty"`
	Job          string    `json:"job,omitempty"`
	InvocationID string    `json:"invocation_id"`
}
