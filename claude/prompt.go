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
// tools, test it and fix it, wait for its last run to end, then reply
// in the thread with the links, the steps and the run's answer.
const WorkflowBuilder = "The question describes a task-ui workflow to build or change. Build it with task-ui's MCP tools as the user: " +
	"look with list_workflows, get_workflow and list_jobs, then create_workflow or update_workflow. " +
	"Prefer prompt steps, which ask Claude with the output of the steps before them, and existing jobs over writing files into a checkout; " +
	"end with a summary step when the answer is a summary of the steps. " +
	"Then run it with run_workflow and a wait, read how each step went and the answer, fix what failed and run it again, at most 3 runs in all. " +
	"A run still going when the wait ends is no failure: wait for it with list_workflow_runs until it ended. " +
	"Once the last run ended, reply in the thread, as the prompt's source says, with the workflow's link, /workflows/<id>, the run's link, /workflows/<id>/runs/<run>, its steps in a short list, " +
	"and the output of its final step verbatim, Claude's answer when it ends with a prompt or a summary. Say you couldn't only for a step that failed."

// RunTriage is how Claude triages a failed scheduled run: the
// instruction beside the thread of a schedule's triage, which asks it to
// find the cause, fix it in the folder when it can, and end with a line
// the scheduler reads as its verdict.
const RunTriage = "This is the triage of a scheduled run that failed. Read the run's output and the job in the folder you are in: the pipeline file, the script or the target the job runs, and what they read. " +
	"Find the root cause. When the fix is inside this folder, in its code, its pipeline file or configuration it may edit, make it, commit it as the rules of this prompt say, and say what changed and where. " +
	"Do not change task-ui's settings, its users or its secrets, and do not run the job again unless the fix is trivially safe. " +
	"End your answer with exactly one line, `Result: resolved`, `Result: not resolved` or `Result: no action`: no action means nothing to fix from here, a cause outside the folder, or a transient one. " +
	"Then one or two lines of what you did."
