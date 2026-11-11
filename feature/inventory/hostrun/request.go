package hostrun

import (
	"github.com/titpetric/task-ui/runner"
)

// Request is what starts the runs of a command on hosts, as a user.
type Request struct {
	// Targets are the hosts, by address: hosts of the inventory, or *
	// for every host task-ui has an ssh login to. None is *.
	Targets []string
	// Command is the shell command each host runs, the runs' args.
	Command string
	// Batch groups the runs under one batch ID, the inventory's Run and
	// a target of * do; a run on one named host is alone without it.
	Batch bool
	// ParentID and ParentKind are what dispatched the runs, as
	// runner.Request has them: a workflow run whose step they are. Empty
	// leaves a batch under its own first run and one run with no parent.
	ParentID   string
	ParentKind string
	// Scheduled, Force and Reason, Cols and Rows go to each run as
	// runner.Request has them.
	Scheduled bool
	Force     bool
	Reason    string
	Cols      int
	Rows      int
}

// parent returns what the runs of a request sit under, with the kind: what
// dispatched them, else the batch's first run once it started.
func (req Request) parent(first *runner.Session) (string, string) {
	switch {
	case req.ParentID != "":
		return req.ParentID, req.ParentKind
	case first == nil:
		return "", ""
	}
	return first.ID(), ""
}
