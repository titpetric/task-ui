package agent

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
)

// Prompt runs Claude Code on the thread in TASK_UI_PROMPT, whose last message is the question,
// in TASK_UI_PROMPT_DIR, or else the workspace's home project, TASK_UI_HOME, or its directory, TASK_UI_WORKSPACE_DIR.
// Claude's reply goes to the terminal and to the file TASK_UI_PROMPT_REPLY names.
// TASK_UI_PROMPT_RUN names a file with the run a thread is about, TASK_UI_PROMPT_WORKFLOW one with a workflow run's steps.
// The system prompt holds TASK_UI_PROMPT_INSTRUCTIONS, the thread TASK_UI_PROMPT_SOURCE to reply in with the task MCP,
// the hosts of TASK_UI_HOSTS and the context; with a source, the final message is the fallback reply.
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
	if a.getenv("CLAUDE_CODE_OAUTH_TOKEN") == "" {
		return errNoToken
	}
	source := a.getenv("TASK_UI_PROMPT_SOURCE")
	prompt := "A prompt from a task-ui folder page. The thread, oldest first; with several messages each is under its author, your earlier answers among them, and the last message is the question:\n\n" +
		thread + "\n\n" + answerLine(source, "the page")
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

	out := a.Stdout
	if path := a.getenv("TASK_UI_PROMPT_REPLY"); path != "" {
		reply, err := os.Create(path)
		if err != nil {
			return err
		}
		defer reply.Close()
		out = io.MultiWriter(a.Stdout, reply)
	}
	return a.ask(ctx, dir, prompt, system(a.getenv("TASK_UI_PROMPT_INSTRUCTIONS"), source, inventory(a.getenv("TASK_UI_HOSTS")), a.context()), out)
}

// answerLine closes a prompt: the final message is what the page shows
// under the question, or, with a source to reply in, the fallback.
func answerLine(source, page string) string {
	if source != "" {
		return "Post the answer as a reply in the thread, as the prompt's source in the system prompt says; your final message is the fallback, in markdown, shown under the question on " + page + " only when you posted no reply."
	}
	return "Your final message is the answer, in markdown, shown under the question on " + page + "."
}

// system joins what goes after Claude's own system prompt, the parts
// that are not empty, a blank line between them.
func system(parts ...string) string {
	kept := parts[:0]
	for _, part := range parts {
		if part != "" {
			kept = append(kept, part)
		}
	}
	return strings.Join(kept, "\n\n")
}

// aboutRun is the prompt of a thread from a run's page: the run, as
// TASK_UI_PROMPT_RUN's file describes it with the end of its output,
// then the thread. Claude commits a fix the question names to main.
func aboutRun(run, thread, source string) string {
	return "A prompt from the page of a task-ui run, about that run. The run:\n\n" +
		strings.TrimSpace(run) + "\n\n" +
		"The question is about this run. If it asks for a fix to the job or the pipeline, make it in the checkout you are in, commit it to main and push it, and say what changed and where. " +
		"Otherwise change nothing. The thread, oldest first; with several messages each is under its author, your earlier answers among them, and the last message is the question:\n\n" +
		thread + "\n\n" + answerLine(source, "the run's page")
}

// aboutWorkflow is the prompt a workflow asks once a run's steps
// ended, or a prompt step asks after the steps before it: the run, as
// TASK_UI_PROMPT_WORKFLOW's file describes it with the end of each
// step's output, then the thread, the prompt first. Claude changes
// nothing the prompt does not name.
func aboutWorkflow(run, thread, source string) string {
	return "A prompt of a task-ui workflow, asked in a run of the workflow once the steps before it ended, about that run. The run:\n\n" +
		strings.TrimSpace(run) + "\n\n" +
		"Change nothing unless the prompt asks for a change; if it asks for a fix, make it in the checkout you are in, commit it to main and push it, and say what changed and where. " +
		"The thread, oldest first; with several messages each is under its author, your earlier answers among them, and the first message is the workflow's prompt:\n\n" +
		thread + "\n\n" + answerLine(source, "the workflow run's page")
}

// inventory tells Claude about the hosts of TASK_UI_HOSTS and how to
// inspect them over ssh, or passes on why there are none, a line without
// tabs; nothing without hosts.
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
