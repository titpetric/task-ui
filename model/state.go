package model

// State is how a run is doing, or how it ended. The values mirror the
// CHECK constraint in schema/run.up.sql.
type State = string

// The states a run passes through and ends in.
const (
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

// States returns every state, running first.
func States() []State {
	return []State{StateRunning, StateOK, StateFailed, StateStopped, StateKilled}
}

// Finished reports whether a state is one a run ended in.
func Finished(state State) bool {
	return state != StateRunning && state != ""
}
