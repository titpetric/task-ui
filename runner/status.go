package runner

import (
	"github.com/titpetric/task-ui/model"
)

// Status is how a run is doing: its state, and the exit code once it
// ended.
type Status struct {
	State    model.State `json:"state"`
	ExitCode int         `json:"exit_code"`
	// Message says why, when the state needs saying: the pull that
	// failed, the command that could not start.
	Message string `json:"message,omitempty"`
}

// Finished reports whether the status is a final one.
func (s Status) Finished() bool {
	return model.Finished(s.State)
}
