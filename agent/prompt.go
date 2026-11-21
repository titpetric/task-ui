package agent

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
)

// Prompt answers the thread in TASK_UI_PROMPT and writes the reply to the terminal and to TASK_UI_PROMPT_REPLY.
// It errors when the thread is empty, no directory resolves, or no token is set. The package doc lists the variables it reads.
func (a *Agent) Prompt(ctx context.Context) error {
	thread := a.getenv("TASK_UI_PROMPT")
	if thread == "" {
		return errors.New("TASK_UI_PROMPT: this job runs from a prompt against a workspace")
	}
	dir := a.getenv("TASK_UI_PROMPT_DIR")
	if dir == "" {
		dir = a.getenv("TASK_UI_HOME")
	}
	if dir == "" {
		dir = a.getenv("TASK_UI_WORKSPACE_DIR")
	}
	if dir == "" {
		return errors.New("TASK_UI_PROMPT_DIR or TASK_UI_WORKSPACE_DIR: this job runs in a folder or a workspace")
	}
	if !a.signedIn() {
		return errNoToken
	}
	source := a.getenv("TASK_UI_PROMPT_SOURCE")
	prompt := ""
	if a.getenv("TASK_UI_CLAUDE_RESUME") != "" {
		prompt = "task-ui restarted while you worked. Continue the same task in this conversation and finish it. " + answerLine(promptPage(a))
	} else {
		prompt = "A prompt from a task-ui folder page. The thread, oldest first; with several messages each is under its author, your earlier answers among them, and the last message is the question:\n\n" +
			thread + "\n\n" + answerLine("the page")
		if path := a.getenv("TASK_UI_PROMPT_RUN"); path != "" {
			run, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			prompt = aboutRun(string(run), thread, source)
		}
		if path := a.getenv("TASK_UI_PROMPT_WORKFLOW"); path != "" {
			run, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			prompt = aboutWorkflow(string(run), thread, source)
		}
	}

	var out io.Writer
	if path := a.getenv("TASK_UI_PROMPT_REPLY"); path != "" {
		reply, err := os.Create(path)
		if err != nil {
			return err
		}
		defer reply.Close()
		out = reply
	}
	return a.ask(ctx, dir, prompt, system(a.getenv("TASK_UI_PROMPT_INSTRUCTIONS"), promptAgent(a), source, inventory(a.getenv("TASK_UI_HOSTS")), a.context()), out)
}

// promptPage names the page a resumed prompt's answer is shown on.
func promptPage(a *Agent) string {
	switch {
	case a.getenv("TASK_UI_PROMPT_WORKFLOW") != "":
		return "the workflow run's page"
	case a.getenv("TASK_UI_PROMPT_RUN") != "":
		return "the run's page"
	}
	return "the page"
}

// promptAgent names the agent a thread's replies are posted as.
func promptAgent(a *Agent) string {
	if name := a.getenv("TASK_UI_PROMPT_AGENT"); name != "" {
		return "You answer this thread as " + name + "."
	}
	return ""
}

// answerLine closes a prompt: its final message is the answer, in markdown.
func answerLine(page string) string {
	return "Your final message is the answer, in markdown, shown under the question on " + page + "."
}

// system joins the non-empty parts with a blank line between them.
func system(parts ...string) string {
	kept := parts[:0]
	for _, part := range parts {
		if part != "" {
			kept = append(kept, part)
		}
	}
	return strings.Join(kept, "\n\n")
}

// aboutRun returns the prompt of a thread from a run's page: run, then thread.
func aboutRun(run, thread, source string) string {
	return "A prompt from the page of a task-ui run, about that run. The run:\n\n" +
		strings.TrimSpace(run) + "\n\n" +
		"The question is about this run. If it asks for a fix to the job or the pipeline, make it in the checkout you are in, commit it to main and push it, and say what changed and where. " +
		"Otherwise change nothing. The thread, oldest first; with several messages each is under its author, your earlier answers among them, and the last message is the question:\n\n" +
		thread + "\n\n" + answerLine("the run's page")
}

// aboutWorkflow returns the prompt a workflow step asks: run, then thread, the workflow's prompt first.
func aboutWorkflow(run, thread, source string) string {
	return "A prompt of a task-ui workflow, asked in a run of the workflow once the steps before it ended, about that run. The run:\n\n" +
		strings.TrimSpace(run) + "\n\n" +
		"Change nothing unless the prompt asks for a change; if it asks for a fix, make it in the checkout you are in, commit it to main and push it, and say what changed and where. " +
		"The thread, oldest first; with several messages each is under its author, your earlier answers among them, and the first message is the workflow's prompt:\n\n" +
		thread + "\n\n" + answerLine("the workflow run's page")
}

// toolset lists tools, one a line, with the `<tool> --help` hint. Empty for no tools.
func toolset(tools string) string {
	if tools == "" {
		return ""
	}
	return "The tools task-ui found in this environment are available to you here, one a line with what it is used for:\n\n" +
		tools + "\n\n" +
		"Run `<tool> --help` for how to use one."
}

// inventory lists hosts with the ssh command that inspects one. A hosts string with no tab is passed through as the reason there are none.
func inventory(hosts string) string {
	switch {
	case hosts == "":
		return ""
	case !strings.Contains(hosts, "\t"):
		return "task-ui's inventory of hosts: " + hosts
	}
	return "task-ui's inventory lists these hosts, one a line, with tabs between the name, the host, the user that signs in and up or down (an empty user is a host task-ui does not reach):\n\n" +
		hosts + "\n\n" +
		"The run's ssh agent holds the keys that sign in, so you may inspect the hosts with\n\n" +
		"    ssh -o BatchMode=yes -o ConnectTimeout=10 -o StrictHostKeyChecking=accept-new <user>@<host> '<command>'\n\n" +
		"such as `uptime`, `df -h /`, `free -h`, `docker ps --format '{{.Names}}\\t{{.Image}}\\t{{.Status}}'`, `uname -r` or `cat /etc/os-release`. " +
		"For a question about all the hosts, run them in parallel. Prefer read-only commands, and change nothing on a host unless the question asks for that change. " +
		"Say which hosts you could not reach."
}
