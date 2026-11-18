package model

// A room takes participants while open and keeps none once ended.
const (
	MeetingOpen  = "open"
	MeetingEnded = "ended"
)

// The kinds of line a room's log holds.
const (
	LineSaid   = "said"
	LineJoin   = "join"
	LineLeave  = "leave"
	LineRejoin = "rejoin"
	LineEnd    = "end"
)

// The states of a summary: working while Claude reads, done when it wrote, failed when it could not.
const (
	SummaryWorking = "working"
	SummaryDone    = "done"
	SummaryFailed  = "failed"
)

// MeetingLimit is the time a room runs for unless it is ended first.
const MeetingLimit = 3600
