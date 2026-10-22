package claude

// RunSummary is how Claude summarises a run, a job's or a workflow's:
// the instruction after the description of the input, for the Summary
// of a run and for the prompt that Run and summarize asks.
const RunSummary = "Summarise the run in one or two short paragraphs: what it did and how it ended, and the cause if it failed. " +
	"Then, only if the output has values worth seeing at a glance (versions, images, counts, durations, failures per step, hosts), a markdown table of them. " +
	"No preamble, no headings, no advice unless it failed, don't explain the obvious. Change nothing."
