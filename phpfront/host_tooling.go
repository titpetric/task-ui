package phpfront

// ToolingScope is one database atkins records to, as atkins reports it. The
// rollups and the charts over it are PHP's.
//
// From and To bound the window as dates, and To is the day atkins made the
// report, which the window of days counts back from.
type ToolingScope struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Database   string `json:"database"`
	Days       int    `json:"days"`
	From       string `json:"from"`
	To         string `json:"to"`

	// Runs is how many commands ran in the window and TotalMs how long they
	// took between them.
	Runs    int64 `json:"runs"`
	TotalMs int64 `json:"total_ms"`

	// Commands are the command keys by total time, Usage a row per tool per
	// day and CommandUsage a row per command per day.
	Commands     []ToolingCommand `json:"commands"`
	Usage        []ToolingUsage   `json:"usage"`
	CommandUsage []ToolingUsage   `json:"command_usage"`

	// Samples are the commands of the window one by one, oldest first, which
	// a chart buckets finer than a day. An atkins that lists none leaves the
	// charts to the daily series.
	Samples []ToolingSample `json:"samples"`
}

// ToolingCommand is one command key and what it has cost, which the table
// behind the panel's toggle lists.
type ToolingCommand struct {
	Command   string `json:"command"`
	Tool      string `json:"tool"`
	Runs      int64  `json:"runs"`
	TotalMs   int64  `json:"total_ms"`
	AverageMs int64  `json:"average_ms"`
	MaxMs     int64  `json:"max_ms"`
	Failures  int64  `json:"failures"`

	// LastRunAt is when it last ran, as Task\Clock reads a stamp, and empty
	// for a command that has not run in the window.
	LastRunAt string `json:"last_run_at"`
}

// ToolingUsage is one name on one day: a tool of the usage series, or a
// command of the command series, which an older atkins does not report. A day
// the name did not run has no row.
type ToolingUsage struct {
	Date    string `json:"date"`
	Command string `json:"command"`
	Tool    string `json:"tool"`
	Runs    int64  `json:"runs"`
	TotalMs int64  `json:"total_ms"`
}

// ToolingSample is one command atkins recorded: when it started, its key and
// the job it ran under, how long it took and how it ended. It is named by its
// key and never by its command line, which can carry a token.
type ToolingSample struct {
	// At is when it started, as Task\Clock reads a stamp, and AtUnix the
	// same moment as a unix time, 0 for a sample that names none.
	//
	// Both, because the two readers want different things: the list of the
	// newest commands prints the stamp, and the charts bucket by the second.
	// A chart that read the second off the text cost a parse per sample,
	// which is a regexp and two crossings into the host for a number the
	// report was built from: 107ms of a 2s /runs.
	At         string `json:"at"`
	AtUnix     int64  `json:"at_unix"`
	Command    string `json:"command"`
	Tool       string `json:"tool"`
	Job        string `json:"job"`
	DurationMs int64  `json:"duration_ms"`
	ExitCode   int64  `json:"exit_code"`
	Failed     bool   `json:"failed"`
}
