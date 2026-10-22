package agent

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
)

// Prompt runs Claude Code on a prompt, TASK_UI_PROMPT, in the directory
// TASK_UI_PROMPT_DIR names, a folder or a directory inside one, or else
// in the workspace's checkout, TASK_UI_WORKSPACE_DIR: Claude reads the
// CLAUDE.md and .claude/ there and above. The prompt is a thread, oldest
// first, each message under its author in a thread of more than one; the
// last is the question. Claude's answer goes to the terminal and to the file
// TASK_UI_PROMPT_REPLY names, which the board replies with. A thread
// from a run's page has the run in the file TASK_UI_PROMPT_RUN names.
func (a *Agent) Prompt(ctx context.Context) error {
	thread := a.getenv("TASK_UI_PROMPT")
	if thread == "" {
		return errors.New("TASK_UI_PROMPT: this job runs from a prompt against a workspace")
	}
	dir := a.getenv("TASK_UI_PROMPT_DIR")
	if dir == "" {
		dir = a.getenv("TASK_UI_WORKSPACE_DIR")
	}
	if dir == "" {
		return errors.New("TASK_UI_PROMPT_DIR or TASK_UI_WORKSPACE_DIR: this job runs in a folder or a workspace")
	}
	if a.getenv("CLAUDE_CODE_OAUTH_TOKEN") == "" {
		return errNoToken
	}
	prompt := "A prompt from a task-ui folder page. The thread, oldest first; with several messages each is under its author, your earlier answers among them, and the last message is the question:\n\n" +
		thread + "\n\n" +
		"Your final message is the answer, in markdown, shown under the question on the page."
	if path := a.getenv("TASK_UI_PROMPT_RUN"); path != "" {
		run, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		prompt = aboutRun(string(run), thread)
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
	return a.ask(ctx, dir, prompt, inventory(a.getenv("TASK_UI_HOSTS")), out)
}

// aboutRun is the prompt of a thread from a run's page: the run, as
// TASK_UI_PROMPT_RUN's file describes it with the end of its output,
// then the thread. Claude commits a fix the question names to main.
func aboutRun(run, thread string) string {
	return "A prompt from the page of a task-ui run, about that run. The run:\n\n" +
		strings.TrimSpace(run) + "\n\n" +
		"The question is about this run. If it asks for a fix to the job or the pipeline, make it in the checkout you are in, commit it to main and push it, and say what changed and where. " +
		"Otherwise change nothing. The thread, oldest first; with several messages each is under its author, your earlier answers among them, and the last message is the question:\n\n" +
		thread + "\n\n" +
		"Your final message is the answer, in markdown, shown under the question on the run's page."
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
