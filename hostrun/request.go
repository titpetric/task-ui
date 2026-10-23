package hostrun

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
	// Scheduled, Force and Reason, Cols and Rows go to each run as
	// runner.Request has them.
	Scheduled bool
	Force     bool
	Reason    string
	Cols      int
	Rows      int
}
