package api

import (
	"github.com/titpetric/task-ui/runner"
)

// InstanceState is how the slots and the backlog stand, with the parent
// environment mode in effect.
type InstanceState struct {
	runner.BacklogState
	ParentEnv string `json:"parent_env"`
}
