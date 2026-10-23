package view

import (
	"time"
)

// ToolingInvocation is one command of the list under the charts: when
// it started, and When as the list says it, its key and job, its colour on the charts, how long it
// took and how it ended.
type ToolingInvocation struct {
	At      time.Time
	When    string
	Command string
	Job     string
	Color   string
	Took    time.Duration
	Failed  bool
	Exit    int64
}
