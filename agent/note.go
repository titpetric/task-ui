package agent

import (
	"context"
	"errors"
	"io"
	"os"
)

// Note runs Claude Code to the end on the Notes board thread in TASK_UI_NOTE, in TASK_UI_WORKSPACE_DIR.
// It fails without the thread or a token, or when it cannot create the file TASK_UI_NOTE_REPLY names.
// Claude's answer goes to the terminal and to that file, which the board replies with.
func (a *Agent) Note(ctx context.Context) error {
	thread := a.getenv("TASK_UI_NOTE")
	if thread == "" {
		return errors.New("TASK_UI_NOTE: this job runs from a note on the Notes board")
	}
	if !a.signedIn() {
		return errNoToken
	}
	author := a.getenv("TASK_UI_NOTE_AUTHOR")
	if author == "" {
		author = "a user"
	}
	prompt := "An instruction on the task-ui Notes board, note " + a.getenv("TASK_UI_NOTE_ID") + " by " + author + ". The thread, oldest first, each message under its author; the last message naming you is the instruction, and your earlier replies are there too:\n\n" +
		thread + "\n\n" +
		"The working directory holds the projects task-ui runs, each a git checkout of its own. Commit what you change to main and push it. Your final message is your reply on the note, in markdown, to its author: what you changed and where, or the answer they asked for. If you will not or cannot do it, do nothing and say why instead."

	out := a.Stdout
	if path := a.getenv("TASK_UI_NOTE_REPLY"); path != "" {
		reply, err := os.Create(path)
		if err != nil {
			return err
		}
		defer reply.Close()
		out = io.MultiWriter(a.Stdout, reply)
	}
	return a.ask(ctx, a.getenv("TASK_UI_WORKSPACE_DIR"), prompt, a.context(), out)
}
