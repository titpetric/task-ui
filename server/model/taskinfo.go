package model

// TaskInfo holds information available about a task, without
// coupling to taskfile structures too much. We need a subset
// of data available for the UI, we can still use the rest
// internally.
//
// This is also the point where exposing vars and cmds becomes
// a security issue. We still assume that users may want to
// run it in less trusted environments, in which case we
// explicitly don't want any information about the contents
// of the taskfile beyond the target names and descriptions.
type TaskInfo struct {
	Task        string // Task holds the task name, e.g. `lint`.
	Description string // Description holds the task description.

	// Flags holds some flags
	Flags TaskFlags

	// History holds history data
	History []HistoryRecord
}

// TaskFlags holds some task flags.
type TaskFlags struct {
	Internal    bool
	Interactive bool
}

// NewTaskInfo converts the taskfile record into our own.
func NewTaskInfo(spec *Task) *TaskInfo {
	return &TaskInfo{
		Task:        spec.Name,
		Description: spec.Description,
		Flags: TaskFlags{
			Internal:    spec.Internal,
			Interactive: spec.Interactive,
		},
	}
}
