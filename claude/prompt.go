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

// WorkflowBuilder is what the prompt of the Workflows page tells
// Claude: build the workflow the question describes with task-ui's
// tools, test it and fix it, then answer with the link, the steps and
// the last output.
const WorkflowBuilder = "The question describes a task-ui workflow to build or change. Build it with task-ui's MCP tools as the user: " +
	"look with list_workflows, get_workflow and list_jobs, then create_workflow or update_workflow. " +
	"Prefer prompt steps, which ask Claude with the output of the steps before them, and existing jobs over writing files into a checkout. " +
	"Then run it with run_workflow and a wait, read how each step went and the answer, fix what failed and run it again, at most 3 runs in all. " +
	"Answer with the workflow's link, /workflows/<id>, its steps in a short list, and the output of its final step verbatim, Claude's answer when it ends with a prompt."
