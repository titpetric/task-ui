package phpfront

// Running is one run in progress, by the key its job is listed under.
type Running struct {
	Key   string `json:"key"`
	ID    string `json:"id"`
	State string `json:"state"`
}

// Flow is one workflow run in progress, by the workflow it belongs to.
type Flow struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id"`
}

// Session is what the runner knows about a run that the table does not yet:
// the state of a transition it has not written, the status it reports, the
// windows attached to it and its place in the queue.
//
// Live is false where the runner holds no session for the ID, and then the
// row is the whole answer. Message is empty where the status says nothing and
// the run waits for nothing, so a page falls back to the column.
type Session struct {
	Live       bool   `json:"live"`
	Attached   int    `json:"attached"`
	Message    string `json:"message"`
	State      string `json:"state"`
	ExitCode   int64  `json:"exit_code"`
	Title      string `json:"title"`
	FinishedAt string `json:"finished_at"`
}

// Screen is the last frame of a run's output as the SVG a page embeds, with
// the lines of its whole output and whether the frame already shows them all.
//
// Lines is 0 and Complete false for a run still going, whose output may still
// scroll; Complete is true for a recording the retention sweep has removed,
// where the frame says so and there is nothing left to expand.
type Screen struct {
	SVG      string `json:"svg"`
	Lines    int    `json:"lines"`
	Complete bool   `json:"complete"`
}

// Rules is config/permissions.yml: the titles in file order, and the path
// lists the file decides. Paths and Checks come over already built, so the
// decision logic is the only part of permission/ that is ported.
type Rules struct {
	Rules       []Rule   `json:"rules"`
	Sections    []Rule   `json:"sections"`
	Permissions []Rule   `json:"permissions"`
	Paths       []string `json:"paths"`
	Checks      []string `json:"checks"`

	// Extras are the rules a section added of its own, which the section's
	// all does not stand for. The key is the path, <section>/<rule>, because
	// that is what a check names and what a policy sets; the title is the
	// rule's own, which the matrix composes with the section's title.
	Extras []Rule `json:"extras"`
}

// Rule is one titled key of the permission file.
type Rule struct {
	Key   string `json:"key"`
	Title string `json:"title"`
}
