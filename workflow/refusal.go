package workflow

// Refusal names the step a start may not run and who may not run it:
// the user who starts the run, empty, or the user the workflow runs as.
// The zero Refusal lets the start go.
type Refusal struct {
	Step string
	User string
}

// String says why the start may not go, "" for a zero Refusal.
func (r Refusal) String() string {
	switch {
	case r.Step == "":
		return ""
	case r.User == "":
		return "you don't have permission to run " + r.Step
	}
	return r.User + " may not run " + r.Step
}
