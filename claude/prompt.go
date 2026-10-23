package claude

// RunSummary is how Claude summarises a run, a job's or a workflow's:
// the instruction after the description of the input, for the Summary
// of a run and for the prompt that Run and summarize asks.
const RunSummary = "Summarise the run in one or two short paragraphs: what it did and how it ended, and the cause if it failed. " +
	"Then, only if the output has values worth seeing at a glance (versions, images, counts, durations, failures per step, hosts), a markdown table of them. " +
	"No preamble, no headings, no advice unless it failed, don't explain the obvious. Change nothing."

// HostsSummary is how Claude summarises the runs of one command on
// hosts, a batch of @host:<host> runs: the short summary of RunSummary,
// with a table a row a host.
const HostsSummary = "The runs are one shell command on several hosts over ssh, a run a host. " +
	"Summarise them in one or two short paragraphs: lead with what needs attention (failures, hosts that differ, warnings, anything near a limit), then a line on what is fine. " +
	"Then a markdown table with a row a host: the host, how its run ended, and the values worth seeing at a glance. " +
	"No preamble, no headings, no advice unless one failed, don't explain the obvious. Change nothing."
