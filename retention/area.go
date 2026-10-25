package retention

// Signins are the expired tokens and sessions, the one area no rule
// governs: removed every cleanup.
const Signins = "signins"

// Area is a row of the Data retention table: a kind of data, what goes
// when its items expire, and whether the items have a size. A fixed
// area has no rule: the row says what happens to it instead.
type Area struct {
	Kind  string
	Title string
	// What says what goes when an item expires, and what stays.
	What string
	// One and Many name an item of the area: run, runs.
	One, Many string
	// Sized says the items have a size: the page shows the bytes beside
	// the count, and a size limit applies unless Unlimited.
	Sized bool
	// Unlimited says a sized area takes no size limit, only an age.
	Unlimited bool
	// Fixed is what the row says in place of a rule; empty for an area
	// with one.
	Fixed string
}

// Areas are the rows of the table in order: the kinds with a rule, the
// fixed area among them. Clean applies them in this order, so run history
// takes a run's recording, summary and threads before the areas after it
// count them.
var Areas = []Area{
	{Kind: Recordings, Title: "Run recordings", One: "recording", Many: "recordings", Sized: true,
		What: "The terminal recording and output of a finished run, the .cast file Replay reads. The run, who started it and how it ended stay."},
	{Kind: Runs, Title: "Run history", One: "run", Many: "runs", Sized: true,
		What: "The rows of finished runs older than the age, with their recordings, summaries and the prompt threads about them. The size is their recordings'."},
	{Kind: Prompts, Title: "Prompt threads", One: "thread", Many: "threads",
		What: "Prompt threads whose last reply is older than the age, with their replies. The runs stay; the notes people write on the board are not touched."},
	{Kind: WorkflowRuns, Title: "Workflow run history", One: "run", Many: "runs",
		What: "The runs of workflows older than the age, with their steps, answers and the threads of their prompts. The workflows stay."},
	{Kind: Artifacts, Title: "Artifact history", One: "version", Many: "versions", Sized: true,
		What: "The earlier versions of the artifact tree's files, which the history keeps when a run or a user writes a file again. The latest version of every file stays."},
	{Kind: SecretReads, Title: "Secret read log", One: "read", Many: "reads",
		What: "Which run got which secret in its environment. The log keeps the last 1000 reads of each secret whatever the rule says."},
	{Kind: Signins, Title: "Expired sign-ins", One: "row", Many: "rows",
		What:  "Tokens and browser sessions past their expiry, which no longer sign in.",
		Fixed: "Removed every cleanup, no setting."},
	{Kind: Tooling, Title: "Tool history", One: "database", Many: "databases", Sized: true, Unlimited: true,
		What: "The runs and commands older than the age in atkins.db of each workspace's home and the instance's own copy: what the tools ran, which atkins prunes and vacuums."},
}

// Limited says a size limit applies to the area.
func (a Area) Limited() bool {
	return a.Sized && !a.Unlimited
}

// AreaOf returns the area of a kind, empty for one the table lacks.
func AreaOf(kind string) Area {
	for _, area := range Areas {
		if area.Kind == kind {
			return area
		}
	}
	return Area{}
}

// Holds reads what an area holds as the page and Clean say it: the count
// and, for a sized area, the bytes.
func (a Area) Holds(u Usage) string {
	if a.Sized {
		return u.Says(a.One, a.Many)
	}
	return u.Name(a.One, a.Many)
}
