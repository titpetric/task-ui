package model

// State is how a run is doing, or how it ended. The values mirror the
// CHECK constraint in schema/run.up.sql.
type State = string

// The states a run passes through and ends in.
const (
	// StateQueued is a run that waits for the run before it of its job
	// to end: a command on a host the host is still running one of.
	StateQueued State = "queued"
	// StateRunning is a run in progress.
	StateRunning State = "running"
	// StateOK is a run that exited with zero.
	StateOK State = "ok"
	// StateFailed is a run that exited with anything else.
	StateFailed State = "failed"
	// StateStopped is a run the user stopped.
	StateStopped State = "stopped"
	// StateKilled is a run stopped from elsewhere: Run force, Kill, or
	// the instance going down.
	StateKilled State = "killed"
)

// QueuedLost is the message of a queued run the instance stopped or
// restarted before the run started.
const QueuedLost = "the instance restarted before it ran"

// States returns every state, queued and running first.
func States() []State {
	return []State{StateQueued, StateRunning, StateOK, StateFailed, StateStopped, StateKilled}
}

// Finished reports whether a state is one a run ended in: neither
// queued nor running.
func Finished(state State) bool {
	return state != StateQueued && state != StateRunning && state != ""
}
