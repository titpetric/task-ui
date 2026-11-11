package tooling

// CommandUsage is one command key on one day: the series a chart
// stacks by command. A day a command did not run has no
// row.
type CommandUsage struct {
	Date    string `json:"date"`
	Command string `json:"command"`
	Tool    string `json:"tool"`
	Runs    int64  `json:"runs"`
	TotalMs int64  `json:"total_ms"`
}
